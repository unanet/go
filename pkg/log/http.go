package log

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

func DecodeBodyFromRequest(r *http.Request) zap.Field {
	if r.Body == nil {
		return zap.String("body", "")
	}
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return zap.Error(err)
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	var b interface{}
	err = json.Unmarshal(buf, &b)
	if err == nil {
		return zap.Any("body", b)
	}

	return zap.String("body", string(buf))
}

func DecodeBodyFromResponse(r *http.Response) zap.Field {
	if r.Body == nil {
		return zap.String("body", "")
	}
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return zap.Error(err)
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	var b interface{}
	err = json.Unmarshal(buf, &b)
	if err == nil {
		return zap.Any("body", b)
	}

	return zap.String("body", string(buf))
}

func DecodeBody(r io.ReadCloser) zap.Field {
	if r == nil {
		return zap.String("body", "")
	}

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

func DecodeHeaderFromRequest(r *http.Request) zap.Field {
	return DecodeHeader(r.Header)
}

func DecodeHeaderFromResponse(r *http.Response) zap.Field {
	return DecodeHeader(r.Header)
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
