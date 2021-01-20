package log

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

func DecodeBody(r io.Reader) zap.Field {
	bodyBytes, err := ioutil.ReadAll(r)
	if err != nil {
		return zap.Error(err)
	}
	var b interface{}
	err = json.Unmarshal(bodyBytes, &b)
	if err == nil {
		return zap.Any("body", b)
	}

	return zap.String("body", string(bodyBytes))
}

func DecodeHeader(h http.Header) zap.Field {
	headers := map[string]string{}
	for k, v := range h {
		if strings.ToLower(k) == "authorization" {
			hash := sha256.New()
			hash.Write([]byte(strings.Join(v, "|")))
			headers[k] = base64.URLEncoding.EncodeToString(hash.Sum(nil))
		} else {
			headers[k] = strings.Join(v, "|")
		}
	}

	return zap.Any("headers", headers)
}
