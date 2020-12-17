package json

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

// jsonDecoder decodes http response JSON into a JSON-tagged struct value.
type jsonDecoder struct {
}

func NewJsonDecoder() *jsonDecoder {
	return &jsonDecoder{}
}

// Decode decodes the Response Body into the value pointed to by v.
// Caller must provide a non-nil v and close the resp.Body.
func (d jsonDecoder) Decode(resp *http.Response, v interface{}) error {
	switch vu := v.(type) {
	case *string:
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil
		}
		*vu = string(bodyBytes)
		return nil
	default:
		err := json.NewDecoder(resp.Body).Decode(v)
		if err == io.EOF {
			err = nil // ignore EOF errors caused by empty response body
		}
		return err
	}
}
