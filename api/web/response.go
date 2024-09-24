package web

import (
	"encoding/json"
	"io"
	"net/http"
)

type Headers interface {
	Headers() http.Header
}

func EncodeJSON(w http.ResponseWriter, v interface{}, code int) error {
	if headers, ok := v.(Headers); ok {
		for k, values := range headers.Headers() {
			for _, v := range values {
				w.Header().Add(k, v)
			}
		}
	}

	if code == http.StatusNoContent {
		w.WriteHeader(code)
		return nil
	}

	var jsonData []byte

	var err error
	switch v := v.(type) {
	case []byte:
		jsonData = v
	case io.Reader:
		jsonData, err = io.ReadAll(v)
	default:
		jsonData, err = json.Marshal(v)
	}

	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.WriteHeader(code)

	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}
