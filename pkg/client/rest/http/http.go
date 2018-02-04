package http

import (
	"bytes"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/api"
	"github.com/Aptomi/aptomi/pkg/api/codec"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"io"
	"io/ioutil"
	"net/http"
)

// Client is the interface for doing HTTP requests that operates using runtime objects
type Client interface {
	GET(path string, expected *runtime.Info) (runtime.Object, error)
	POST(path string, expected *runtime.Info, body runtime.Object) (runtime.Object, error)
	POSTSlice(path string, expected *runtime.Info, body []runtime.Object) (runtime.Object, error)
	DELETESlice(path string, expected *runtime.Info, body []runtime.Object) (runtime.Object, error)
}

type httpClient struct {
	contentType *codec.ContentTypeHandler
	http        *http.Client
	cfg         *config.Client
}

// NewClient returns implementation of
func NewClient(cfg *config.Client) Client {
	client := &http.Client{
		Timeout: cfg.HTTP.Timeout,
	}
	contentTypeHandler := codec.NewContentTypeHandler(runtime.NewRegistry().Append(api.Objects...))

	return &httpClient{contentTypeHandler, client, cfg}
}

func (client *httpClient) GET(path string, expected *runtime.Info) (runtime.Object, error) {
	return client.request(http.MethodGet, path, expected, nil)
}

func (client *httpClient) POST(path string, expected *runtime.Info, body runtime.Object) (runtime.Object, error) {
	var bodyData io.Reader

	if body != nil {
		data, err := client.contentType.GetCodecByContentType(codec.Default).EncodeOne(body)
		if err != nil {
			return nil, fmt.Errorf("error while encoding body for post request: %s", err)
		}
		bodyData = bytes.NewBuffer(data)
	}

	return client.request(http.MethodPost, path, expected, bodyData)
}

func (client *httpClient) POSTSlice(path string, expected *runtime.Info, body []runtime.Object) (runtime.Object, error) {
	var bodyData io.Reader

	if body != nil {
		data, err := client.contentType.GetCodecByContentType(codec.Default).EncodeMany(body)
		if err != nil {
			return nil, fmt.Errorf("error while encoding body for post request: %s", err)
		}
		bodyData = bytes.NewBuffer(data)
	}

	return client.request(http.MethodPost, path, expected, bodyData)
}

func (client *httpClient) DELETESlice(path string, expected *runtime.Info, body []runtime.Object) (runtime.Object, error) {
	var bodyData io.Reader

	if body != nil {
		data, err := client.contentType.GetCodecByContentType(codec.Default).EncodeMany(body)
		if err != nil {
			return nil, fmt.Errorf("error while encoding body for delete request: %s", err)
		}
		bodyData = bytes.NewBuffer(data)
	}

	return client.request(http.MethodDelete, path, expected, bodyData)
}

func (client *httpClient) request(method string, path string, expected *runtime.Info, body io.Reader) (runtime.Object, error) {
	req, err := http.NewRequest(method, client.cfg.API.URL()+path, body)
	if err != nil {
		return nil, err
	}

	if len(client.cfg.Auth.Token) > 0 {
		req.Header.Set("Authorization", "Bearer "+client.cfg.Auth.Token)
	}
	req.Header.Set("Content-Type", codec.Default)
	req.Header.Set("User-Agent", "aptomictl")

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint: errcheck

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error while reading bytes from response Body: %s", err)
	}

	if len(respData) == 0 {
		return nil, fmt.Errorf("empty response")
	}

	obj, err := client.contentType.GetCodec(resp.Header).DecodeOne(respData)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshalling response: %s", err)
	}

	if obj.GetKind() == api.ServerErrorObject.Kind {
		serverErr, ok := obj.(*api.ServerError)
		if !ok {
			return nil, fmt.Errorf("server error, but it couldn't be casted to api.ServerError")
		}

		return nil, fmt.Errorf("server error: %s", serverErr.Error)
	}

	if expected != nil && obj.GetKind() != expected.Kind {
		return nil, fmt.Errorf("received object kind %s doesn't match expected %s", obj.GetKind(), expected.Kind)
	}

	return obj, nil
}
