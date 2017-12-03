package api

import (
	"fmt"
	"github.com/Aptomi/aptomi/pkg/lang"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"net/http"
)

func (api *coreAPI) readLang(request *http.Request) []lang.Base {
	result := make([]lang.Base, 0)

	exists := make(map[string]bool, len(result))
	for _, obj := range api.contentType.Read(request) {
		langObj, ok := obj.(lang.Base)

		if !ok {
			panic(fmt.Sprintf("Trying to read lang objects while non-lang ones found: %s", obj.GetKind()))
		}

		key := runtime.KeyForStorable(langObj)
		if exists[key] {
			panic(fmt.Sprintf("Duplicate objects with key %s detected in the request", key))
		}
		exists[key] = true

		result = append(result, langObj)
	}

	return result
}
