package main

import (
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Proxy struct {
	URL          *url.URL
	ShouldInject func(resp *http.Response) bool
	Inject       string
	Do           func(req *http.Request) (*http.Response, error)
	ShouldRetry  func(req *http.Request) bool
	RetryDelay   time.Duration
	ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)
}

func NewProxy(dst string) (*Proxy, error) {
	u, err := url.Parse(dst)
	if err != nil {
		return nil, err
	}
	return &Proxy{
		URL: u,
		Do:  http.DefaultClient.Do,
		ShouldInject: func(resp *http.Response) bool {
			mediaType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type"))
			return resp.Request.Header.Get("Sec-Fetch-Dest") == "document" &&
				mediaType == "text/html"
		},
		RetryDelay: 5 * time.Millisecond,
		ShouldRetry: func(req *http.Request) bool {
			return req.Header.Get("Sec-Fetch-Dest") == "document"
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Println(err)
		},
	}, nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, src *http.Request) {
	req := p.request(src)
	resp, err := p.Do(req)
	count := 0
	for err != nil && p.ShouldRetry(req) {
		count++
		time.Sleep(p.RetryDelay)
		resp, err = p.Do(req)
	}
	if err != nil {
		p.ErrorHandler(w, req, err)
		return
	}
	defer resp.Body.Close()
	p.respond(w, resp)
}

func (p Proxy) request(src *http.Request) *http.Request {
	req := src.Clone(src.Context())

	req.RequestURI = ""
	if p.URL.Scheme != "" {
		req.URL.Scheme = p.URL.Scheme
	}
	if p.URL.Host != "" {
		req.URL.Host = p.URL.Host
	}

	return req
}

func (p Proxy) respond(dst http.ResponseWriter, src *http.Response) {
	for key, values := range src.Header {
		for _, value := range values {
			dst.Header().Add(key, value)
		}
	}
	if p.ShouldInject(src) {
		src.Body = struct {
			io.Reader
			io.Closer
		}{
			io.MultiReader(src.Body, strings.NewReader(p.Inject)),
			src.Body,
		}
		if src.ContentLength > 0 {
			length := src.ContentLength + int64(len(p.Inject))
			dst.Header().Set("Content-Length", strconv.FormatInt(length, 10))
		}
	}
	dst.WriteHeader(src.StatusCode)
	io.Copy(dst, src.Body)
}
