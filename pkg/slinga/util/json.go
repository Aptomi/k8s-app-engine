package util

import (
	"encoding/json"
	"fmt"
	"io"
)

func WriteJSON(w io.Writer, obj interface{}) error {
	res, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	fmt.Fprint(w, string(res))

	return nil
}
