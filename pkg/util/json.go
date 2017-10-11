package util

import (
	"encoding/json"
	"fmt"
	"io"
)

// WriteJSON writes an object marshalled into JSON into a given writer
func WriteJSON(w io.Writer, obj interface{}) error {
	res, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(w, string(res))
	return err
}
