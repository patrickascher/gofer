// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//TODO test localizer, files
package context_test

import (
	ctx "context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/patrickascher/gofer/controller/context"
	"github.com/patrickascher/gofer/router"
	"github.com/patrickascher/gofer/router/middleware/jwt"
	"github.com/stretchr/testify/assert"
)

func newRequest(r *http.Request) *context.Context {
	w := httptest.NewRecorder()
	return context.New(w, r)
}

func TestRequest_Body(t *testing.T) {
	asserts := assert.New(t)
	r := &http.Request{Body: ioutil.NopCloser(strings.NewReader("text"))}

	// read request body
	req := newRequest(r).Request
	asserts.Equal([]byte("text"), req.Body())
	asserts.Equal([]byte("text"), req.Body())

	// set request body
	req.SetBody([]byte("text2"))
	asserts.Equal([]byte("text2"), req.Body())
}

func TestRequest_Localizer(t *testing.T) {
	// TODO redo
}

func TestRequest_Pattern(t *testing.T) {
	asserts := assert.New(t)
	r := &http.Request{}

	// read the pattern without value
	req := newRequest(&http.Request{}).Request
	asserts.Equal("", req.Pattern())

	// read the pattern with value
	req = newRequest(r.WithContext(ctx.WithValue(r.Context(), router.PATTERN, "/:lang/:user"))).Request
	asserts.Equal("/:lang/:user", req.Pattern())
}

func TestRequest_HTTPRequest(t *testing.T) {
	asserts := assert.New(t)
	r := &http.Request{}

	// read the raw HTTPRequest
	req := newRequest(r).Request
	asserts.Equal(r, req.HTTPRequest())
}

func TestRequest_JWTClaim(t *testing.T) {
	asserts := assert.New(t)
	r := &http.Request{}

	// JWT was not set.
	req := newRequest(r).Request
	asserts.Equal(nil, req.JWTClaim())

	// JWT is set.
	req = newRequest(r.WithContext(ctx.WithValue(r.Context(), jwt.CLAIM, "claim"))).Request
	asserts.Equal("claim", req.JWTClaim().(string))
}

func TestRequest_Method(t *testing.T) {
	asserts := assert.New(t)
	req := newRequest(&http.Request{Method: http.MethodGet}).Request

	// Method function
	asserts.Equal(http.MethodGet, req.Method())

	// Is function
	asserts.True(req.Is(http.MethodGet))
	asserts.False(req.Is(http.MethodPost))

	// IsMethod functions
	asserts.True(req.IsGet())
	asserts.False(req.IsPost())
	req = newRequest(&http.Request{Method: http.MethodPost}).Request
	asserts.True(req.IsPost())
	req = newRequest(&http.Request{Method: http.MethodPut}).Request
	asserts.True(req.IsPut())
	req = newRequest(&http.Request{Method: http.MethodPatch}).Request
	asserts.True(req.IsPatch())
	req = newRequest(&http.Request{Method: http.MethodDelete}).Request
	asserts.True(req.IsDelete())
}

func TestRequest_IsSecure(t *testing.T) {
	asserts := assert.New(t)

	header := http.Header{}
	header["X-Forwarded-Proto"] = []string{"https"}
	req := newRequest(&http.Request{Header: header}).Request
	asserts.True(req.IsSecure())

	header = http.Header{}
	header["X-Forwarded-Proto"] = []string{"http"}
	req = newRequest(&http.Request{Header: header}).Request
	asserts.False(req.IsSecure())
}

func TestRequest_IP(t *testing.T) {
	asserts := assert.New(t)

	// X-Forwarded IP.
	header := http.Header{}
	header["X-Forwarded-For"] = []string{"192.168.2.1"}
	req := newRequest(&http.Request{Header: header}).Request
	asserts.Equal("192.168.2.1", req.IP())

	// Remote Addr with port.
	req = newRequest(&http.Request{}).Request
	req.HTTPRequest().RemoteAddr = "192.168.2.4:8080"
	asserts.Equal("192.168.2.4", req.IP())

	// Remote Addr without port.
	req = newRequest(&http.Request{}).Request
	req.HTTPRequest().RemoteAddr = "192.168.2.4"
	asserts.Equal("192.168.2.4", req.IP())
}

func TestRequest_Proxy(t *testing.T) {
	asserts := assert.New(t)

	// existing X-Forwarted-For header
	header := http.Header{}
	header["X-Forwarded-For"] = []string{"192.168.2.1"}
	req := newRequest(&http.Request{Header: header}).Request
	asserts.Equal([]string{"192.168.2.1"}, req.Proxy())

	// header X-Forwarted-For is not set
	header = http.Header{}
	req = newRequest(&http.Request{Header: header}).Request
	asserts.Equal([]string{}, req.Proxy())
}

func TestRequest_Scheme(t *testing.T) {
	asserts := assert.New(t)

	// X-Forwarded-Proto header https
	header := http.Header{}
	header["X-Forwarded-Proto"] = []string{"https"}
	req := newRequest(&http.Request{Header: header}).Request
	asserts.Equal("https", req.Scheme())

	// X-Forwarded-Proto header http
	header = http.Header{}
	header["X-Forwarded-Proto"] = []string{"http"}
	req = newRequest(&http.Request{Header: header}).Request
	asserts.Equal("http", req.Scheme())

	// URL with schema
	adr, _ := url.Parse("https://test.com:8043/user?q=dotnet")
	req = newRequest(&http.Request{TLS: nil, URL: adr}).Request
	asserts.Equal("https", req.Scheme())

	//TLS
	adr, _ = url.Parse("https://test.com:8043/user?q=dotnet")
	adr.Scheme = ""
	req = newRequest(&http.Request{TLS: &tls.ConnectionState{}, URL: adr}).Request
	asserts.Equal("https", req.Scheme())

	// no TLS
	adr, _ = url.Parse("https://test.com:8043/user?q=dotnet")
	adr.Scheme = ""
	req = newRequest(&http.Request{TLS: nil}).Request
	asserts.Equal("http", req.Scheme())
}

func TestRequest_HostDomain(t *testing.T) {
	asserts := assert.New(t)

	// localhost
	req := newRequest(&http.Request{}).Request
	asserts.Equal("localhost", req.Host())

	// host without port
	req = newRequest(&http.Request{Host: "example.com"}).Request
	asserts.Equal("example.com", req.Host())

	// host with port
	req = newRequest(&http.Request{Host: "example.com:8080"}).Request
	asserts.Equal("example.com", req.Host())

	// domain with port
	req = newRequest(&http.Request{Host: "example.com:8080"}).Request
	asserts.Equal("example.com", req.Domain())
}

func TestRequest_Protocol(t *testing.T) {
	asserts := assert.New(t)

	// HTTP/2
	req := newRequest(&http.Request{Proto: "HTTP/2"}).Request
	asserts.Equal("HTTP/2", req.Protocol())

	// HTTP/1.1
	req = newRequest(&http.Request{Proto: "HTTP/1.1"}).Request
	asserts.Equal("HTTP/1.1", req.Protocol())
}

func TestRequest_URI(t *testing.T) {
	asserts := assert.New(t)
	adr, _ := url.Parse("https://example.com:8080/user?q=golang#test")
	req := newRequest(&http.Request{URL: adr, Host: "example.com:8080", RequestURI: "/user?q=golang#test"}).Request
	asserts.Equal("/user?q=golang#test", req.URI())
}

func TestRequest_URL(t *testing.T) {
	asserts := assert.New(t)
	adr, _ := url.Parse("https://example.com:8080/user?q=golang#test")
	req := newRequest(&http.Request{URL: adr, Host: "example.com:8080", RequestURI: "/user?q=golang#test"}).Request
	asserts.Equal("/user", req.URL())
}

func TestRequest_FullURL(t *testing.T) {
	asserts := assert.New(t)
	adr, _ := url.Parse("https://example.com:8080/user?q=golang#test")
	req := newRequest(&http.Request{URL: adr, Host: "example.com:8080", RequestURI: "/user?q=golang#test"}).Request
	asserts.Equal("https://example.com:8080/user?q=golang#test", req.FullURL())
}

func TestRequest_Site(t *testing.T) {
	asserts := assert.New(t)
	adr, _ := url.Parse("https://example.com:8080/user?q=golang#test")
	req := newRequest(&http.Request{URL: adr, Host: "example.com:8080", RequestURI: "/user?q=golang#test"}).Request
	asserts.Equal("https://example.com", req.Site())
}

func TestRequest_Port(t *testing.T) {
	asserts := assert.New(t)
	adr, _ := url.Parse("https://example.com:8080/user?q=golang#test")
	req := newRequest(&http.Request{URL: adr, Host: "example.com:8080", RequestURI: "/user?q=golang#test"}).Request
	asserts.Equal(8080, req.Port())

	asserts = assert.New(t)
	adr, _ = url.Parse("https://example.com/user?q=golang#test")
	req = newRequest(&http.Request{URL: adr, Host: "example.com", RequestURI: "/user?q=golang#test"}).Request
	asserts.Equal(80, req.Port())
}

func TestRequest_Referer(t *testing.T) {
	asserts := assert.New(t)

	header := http.Header{}
	header["Referer"] = []string{"GoBrowser"}
	req := newRequest(&http.Request{Header: header}).Request
	asserts.Equal("GoBrowser", req.Referer())

	req = newRequest(&http.Request{}).Request
	asserts.Equal("", req.Referer())
}

func TestRequest_parseRouterParams(t *testing.T) {
	asserts := assert.New(t)

	w := httptest.NewRecorder()
	header := http.Header{}
	header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	form := url.Values{}
	form.Add("username", "John Doe")
	r := httptest.NewRequest("GET", "https://test.com:8043/user?id=1#test", strings.NewReader(form.Encode()))
	r = r.WithContext(ctx.WithValue(r.Context(), router.PARAMS, map[string][]string{"lang": {"de"}}))
	r.Header = header
	req := context.New(w, r).Request

	param, err := req.Param("lang")
	asserts.NoError(err)
	asserts.Equal([]string{"de"}, param)

	param, err = req.Param("id")
	asserts.NoError(err)
	asserts.Equal([]string{"1#test"}, param)
}

func TestRequest_parseGet(t *testing.T) {
	asserts := assert.New(t)

	w := httptest.NewRecorder()
	header := http.Header{}
	header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	form := url.Values{}
	form.Add("username", "John Doe")
	r := httptest.NewRequest("GET", "https://test.com:8043/user?id=1#test", strings.NewReader(form.Encode()))
	r.Header = header
	req := context.New(w, r)

	// get params - will trigger parse
	params, err := req.Request.Params()
	asserts.NoError(err)
	asserts.Equal(1, len(params))
	//recall - render should not be called twice
	_, err = req.Request.Params()
	asserts.NoError(err)
	asserts.Equal(1, len(params))

	assert.Equal(t, map[string][]string{"id": {"1#test"}}, params)

	// param id
	param, err := req.Request.Param("id")
	asserts.NoError(err)
	asserts.Equal([]string{"1#test"}, param)

	// param does not exist
	param, err = req.Request.Param("password")
	asserts.Error(err)
	asserts.Nil(param)
	asserts.Equal(err, fmt.Errorf(context.ErrParam, "password"))
}

func TestRequest_parsePost(t *testing.T) {
	asserts := assert.New(t)

	w := httptest.NewRecorder()
	header := http.Header{}
	header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	form := url.Values{}
	form.Add("username", "John Doe")
	r := httptest.NewRequest("POST", "https://test.com:8043/user?id=1#test", strings.NewReader(form.Encode()))
	r.Header = header
	req := context.New(w, r)

	// get params - will trigger parse
	params, err := req.Request.Params()
	asserts.NoError(err)
	asserts.Equal(1, len(params))
	//recall - render should not be called twice
	_, err = req.Request.Params()
	asserts.NoError(err)
	asserts.Equal(1, len(params))

	assert.Equal(t, map[string][]string{"username": {"John Doe"}}, params)

	// param username
	param, err := req.Request.Param("username")
	asserts.NoError(err)
	asserts.Equal([]string{"John Doe"}, param)

	// param does not exist
	param, err = req.Request.Param("password")
	asserts.Error(err)
	asserts.Nil(param)
	asserts.Equal(err, fmt.Errorf(context.ErrParam, "password"))
}
