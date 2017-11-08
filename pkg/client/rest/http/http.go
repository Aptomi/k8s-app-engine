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
	"time"
)

type Client interface {
	GET(path string, expected *runtime.Info) (runtime.Object, error)
	POST(path string, expected *runtime.Info, body runtime.Object) (runtime.Object, error)
	POSTSlice(path string, expected *runtime.Info, body []runtime.Object) (runtime.Object, error)
}

type httpClient struct {
	contentType *codec.ContentTypeHandler
	http        *http.Client
	cfg         *config.Client
}

func NewClient(cfg *config.Client) Client {
	client := &http.Client{
		// todo make configurable
		Timeout: 5 * time.Second,
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

func (client *httpClient) request(method string, path string, expected *runtime.Info, body io.Reader) (runtime.Object, error) {
	req, err := http.NewRequest(method, client.cfg.API.URL()+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Username", client.cfg.Auth.Username)
	req.Header.Set("Content-Type", codec.Default)
	req.Header.Set("User-Agent", "aptomictl")

	fmt.Println("Request:", req)

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint: errcheck

	// todo(slukjanov): process response - check status and print returned data
	fmt.Println("Response:", resp)

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Error while reading bytes from response Body: %s", err))
	}

	// todo bad logging
	fmt.Println("Response data:\n" + string(respData))

	obj, err := client.contentType.GetCodec(resp.Header).DecodeOne(respData)
	if err != nil {
		panic(fmt.Sprintf("Error while unmarshalling response: %s", err))
	}

	if expected != nil && obj.GetKind() != expected.Kind {
		// todo handle
		panic("very bad")
	}

	return obj, nil
}
