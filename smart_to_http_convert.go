// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// HTTP conversions for smart http

package http

import (
	"net/http"
)

type httpHandler struct {
	hd Handler
}

func (h httpHandler) ServeHTTP(rw http.ResponseWriter, rq *http.Request) {
	h.hd.ServeHTTP(get_non_http_RW(rw), get_non_http_RQ(rq))
	//FIXME - Headers need to be reflected
}

//get http Handler from local Handler
func gethttpHandler(h Handler) http.Handler {
	return httpHandler{hd: h}
}

type non_httpResponseWriter struct {
	rw http.ResponseWriter
}

func (r non_httpResponseWriter) Header() Header {
	h := r.rw.Header()
	s := Header{}
	for name, values := range h {
		for _, value := range values {
			s.Add(name, value)
		}
	}
	return s
}
func (r non_httpResponseWriter) Write(b []byte) (int, error) {
	return r.rw.Write(b)
}
func (r non_httpResponseWriter) WriteHeader(i int) {
	r.rw.WriteHeader(i)
}

//takes a http ResponseWriter and return a local ResponseWriter
func get_non_http_RW(r http.ResponseWriter) ResponseWriter {
	return non_httpResponseWriter{rw: r}
}

//takes a http Request and return a local Request
func get_non_http_RQ(rq *http.Request) *Request {
	//FIXME - Request needs to be copied
	z := new(Request)
	return z
}
