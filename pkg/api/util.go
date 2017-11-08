package api

import (
	"github.com/Aptomi/aptomi/pkg/lang"
	log "github.com/Sirupsen/logrus"
	"net/http"
)

func (api *coreAPI) readLang(request *http.Request) []lang.Base {
	result := make([]lang.Base, 0)

	for _, obj := range api.contentType.Read(request) {
		langObj, ok := obj.(lang.Base)

		if !ok {
			// todo replace with error?
			log.Panicf("Trying to read lang objects while non-lang ones found: %s", obj.GetKind())
		}

		result = append(result, langObj)
	}

	return result
}
