package codec

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/codec/yaml"
)

const (
	// Default is the default content type
	Default = YAML

	// YAML is the yaml codec content type
	YAML = "application/yaml"

	// JSON is the json codec content type
	JSON = "application/json"
)

// ContentTypeHandler is a helper for working with Content-Type header and doing read/write for http requests/response
type ContentTypeHandler struct {
	codecs map[string]runtime.Codec
}

// NewContentTypeHandler returns instance of ContentTypeHandler for provided runtime registry
func NewContentTypeHandler(reg *runtime.Types) *ContentTypeHandler {
	codecs := make(map[string]runtime.Codec)
	codecs[YAML] = yaml.NewCodec(reg)
	codecs[JSON] = yaml.NewJSONCodec(reg)

	return &ContentTypeHandler{codecs: codecs}
}

// GetCodecByContentType returns runtime codec for provided content type that should be used
func (handler *ContentTypeHandler) GetCodecByContentType(contentType string) runtime.Codec {
	codec, exist := handler.codecs[contentType]
	if codec == nil || !exist {
		return handler.codecs[Default]
	}

	return codec
}

// GetCodec returns runtime codec for specified http headers based on the content type
func (handler *ContentTypeHandler) GetCodec(header http.Header) runtime.Codec {
	contentType := header.Get("Content-Type")
	if len(contentType) == 0 {
		contentType = Default
	}

	return handler.GetCodecByContentType(contentType)
}

// GetContentType returns content type for provided http headers
func (handler *ContentTypeHandler) GetContentType(header http.Header) string {
	contentType := header.Get("Content-Type")
	if len(contentType) == 0 {
		contentType = Default
	}
	if _, exist := handler.codecs[contentType]; !exist {
		contentType = Default
	}

	return contentType
}

// ReadOne runtime object from the provided request using correct content type (taken from request)
func (handler *ContentTypeHandler) ReadOne(request *http.Request) runtime.Object {
	objects := handler.Read(request)
	if len(objects) != 1 {
		panic(fmt.Sprintf("Expected 1 but read %d from request", len(objects)))
	}

	return objects[0]
}

// Read runtime object(s) from the provided request using correct content type (taken from the request)
func (handler *ContentTypeHandler) Read(request *http.Request) []runtime.Object {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		panic(fmt.Sprintf("Error while reading bytes from request Body: %s", err))
	}

	objects, err := handler.GetCodec(request.Header).DecodeOneOrMany(body)
	if err != nil {
		// todo response with some bad request status code
		panic(fmt.Sprintf("Error decoding policy update request: %s", err))
	}

	return objects
}

// WriteOne runtime object into the provided response writer using correct content type (taken from provided request)
// with default http status (200 OK)
func (handler *ContentTypeHandler) WriteOne(writer http.ResponseWriter, request *http.Request, body runtime.Object) {
	handler.WriteOneWithStatus(writer, request, body, http.StatusOK)
}

// WriteOneWithStatus runtime object into the provided response writer using correct content type (taken from provided request)
// with specified http status
func (handler *ContentTypeHandler) WriteOneWithStatus(writer http.ResponseWriter, request *http.Request, body runtime.Object, status int) {
	writer.Header().Set("Content-Type", handler.GetContentType(request.Header))
	writer.WriteHeader(status)

	if body != nil {
		data, err := handler.GetCodec(request.Header).EncodeOne(body)
		if err != nil {
			panic(fmt.Sprintf("Error while encoding body of kind %s: %s", body.GetKind(), err))
		}

		_, wErr := fmt.Fprint(writer, string(data))
		if wErr != nil {
			panic(fmt.Sprintf("Error while writing body: %s", wErr))
		}
	}
}

// WriteMany runtime objects into the provided response writer using correct content type (taken from provided request)
// with default http status (200 OK)
func (handler *ContentTypeHandler) WriteMany(writer http.ResponseWriter, request *http.Request, body []runtime.Object) {
	handler.WriteManyWithStatus(writer, request, body, http.StatusOK)
}

// WriteManyWithStatus runtime objects into the provided response writer using correct content type (taken from provided request)
// with specified http status
func (handler *ContentTypeHandler) WriteManyWithStatus(writer http.ResponseWriter, request *http.Request, body []runtime.Object, status int) {
	writer.Header().Set("Content-Type", handler.GetContentType(request.Header))
	writer.WriteHeader(status)

	if body != nil {
		data, err := handler.GetCodec(request.Header).EncodeMany(body)
		if err != nil {
			if len(body) > 0 {
				panic(fmt.Sprintf("Error while encoding body of kind %s: %s", body[0].GetKind(), err))
			} else {
				panic(fmt.Sprintf("Error while encoding empty list: %s", err))
			}
		}

		_, wErr := fmt.Fprint(writer, string(data))
		if wErr != nil {
			panic(fmt.Sprintf("Error while writing body: %s", wErr))
		}
	}
}
