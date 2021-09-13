package httpRequester

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
)

type GetMessage func(url string) string
type PostMessage func(url string, data url.Values) string

func readResponse(body io.ReadCloser) string {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(body)
	if err != nil {
		log.Warnln(err)
		return ""
	}

	return buf.String()
}

func Get(url string) string {
	log.Traceln("sending GET http request to:", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Warnln(err)
		return ""
	}

	return readResponse(resp.Body)
}

func Post(url string, data url.Values) string {
	log.Traceln("sending POST http request to:", url)
	log.Traceln("POST data:", data)
	resp, err := http.PostForm(url, data)
	if err != nil {
		log.Warnln(err)
		return ""
	}

	return readResponse(resp.Body)
}
