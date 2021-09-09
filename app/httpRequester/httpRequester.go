package httpRequester

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
)

type GetMessage func(url string) string
type PostMessage func(url string, data url.Values) string

type Requester interface {
	Get(url string) string
	Post(url string, data url.Values) string
}

func readResponse(body io.ReadCloser) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)

	return buf.String()
}

func Get(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	return readResponse(resp.Body)
}

func Post(url string, data url.Values) string {
	resp, err := http.PostForm(url, data)
	if err != nil {
		log.Fatal(err)
	}

	return readResponse(resp.Body)
}
