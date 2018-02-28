package proto

import (
	"net/url"
	"net/http"
	"io"
)

type Client interface {
    Do(req *http.Request) (*http.Response, error)
	Get(url string) (resp *http.Response, err error)
    Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
    PostForm(url string, data url.Values) (resp *http.Response, err error)
    Head(url string) (resp *http.Response, err error)
}