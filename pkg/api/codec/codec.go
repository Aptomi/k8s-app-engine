package codec

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/Aptomi/aptomi/pkg/runtime/codec/yaml"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

const (
	Default = YAML
	YAML    = "application/yaml"
	JSON    = "application/json"
)

type ContentTypeHandler struct {
	codecs map[string]runtime.Codec
}

func NewContentTypeHandler(reg *runtime.Registry) *ContentTypeHandler {
	codecs := make(map[string]runtime.Codec)
	codecs[YAML] = yaml.NewCodec(reg)
	codecs[JSON] = yaml.NewJSONCodec(reg)

	return &ContentTypeHandler{codecs}
}

func (handler *ContentTypeHandler) GetCodecByContentType(contentType string) runtime.Codec {
	codec := handler.codecs[contentType]
	if codec == nil {
		log.Panicf("Codec not found for content type: %s", contentType)
	}

	return codec
}

func (handler *ContentTypeHandler) GetCodec(header http.Header) runtime.Codec {
	contentType := header.Get("Content-Type")
	if len(contentType) == 0 {
		contentType = Default
	}

	return handler.GetCodecByContentType(contentType)
}

func (handler *ContentTypeHandler) GetContentType(header http.Header) string {
	contentType := header.Get("Content-Type")
	if len(contentType) == 0 {
		contentType = "application/yaml"
	}

	return contentType
}

func (handler *ContentTypeHandler) Read(request *http.Request) []runtime.Object {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Panicf("Error while reading bytes from request Body: %s", err)
	}

	objects, err := handler.GetCodec(request.Header).DecodeOneOrMany(body)
	if err != nil {
		// todo response with some bad request status code
		log.Panicf("Error decoding policy update request: %s", err)
	}

	return objects
}

func (handler *ContentTypeHandler) Write(writer http.ResponseWriter, request *http.Request, body runtime.Object) {
	handler.WriteStatus(writer, request, body, http.StatusOK)
}

func (handler *ContentTypeHandler) WriteStatus(writer http.ResponseWriter, request *http.Request, body runtime.Object, status int) {
	data, err := handler.GetCodec(request.Header).EncodeOne(body)
	if err != nil {
		// todo should we log such errors?
		log.Panicf("Error while encoding body of kind: ", body.GetKind())
	}

	writer.Header().Set("Content-Type", handler.GetContentType(request.Header))
	writer.WriteHeader(status)

	_, wErr := fmt.Fprint(writer, string(data))
	if wErr != nil {
		log.Panicf("Error while writing body: %s", wErr)
	}
}
