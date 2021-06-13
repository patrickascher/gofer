// Copyright (c) 2021 Patrick Ascher <development@fullhouse-productions.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package context todos:
// TODO parse files, localizer
// TODO body/file size limit
package context

import (
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/patrickascher/gofer/locale/translation"
	"github.com/patrickascher/gofer/router/middleware/jwt"
)

// Error messages.
var (
	ErrParam = "context: the param %#v does not exist"
)

// Localizer of the request.
// Its added by the server.Translation.
// The controller can not call the server package (import cycle), thats why its here.
var (
	DefaultLang string
)

// Request struct.
type Request struct {
	r *http.Request

	body   []byte
	params map[string][]string
	files  map[string][]*multipart.FileHeader

	locale translation.Locale
}

// Body reads the raw body data.
func (r *Request) Body() []byte {
	if r.body == nil {
		b, err := ioutil.ReadAll(r.r.Body)
		if err == nil {
			r.body = b
		}
	}
	return r.body
}

// SetBody for manipulations.
func (r *Request) SetBody(body []byte) {
	r.body = body
}

// Locale is used to translate message ids in the controller.
func (r *Request) Locale() translation.Locale {
	return r.locale
}

// Pattern returns the router url pattern.
// The pattern will be checked by the request context with the key "router_pattern".
// If the pattern is not set, an empty string will return.
//		Example: http://example.com/user/1
// 		/user/:id
func (r *Request) Pattern() string {
	if p := r.HTTPRequest().Context().Value("router_pattern"); p != nil { // used string instead of router.PATTERN because of dependency cycle.
		return p.(string)
	}
	return ""
}

// HTTPRequest returns the original *http.Request.
func (r *Request) HTTPRequest() *http.Request {
	return r.r
}

// JWTClaim is a helper to return the claim.
// The claim will be checked by the request context with the key "JWT".
// Nil will return if it was not set.
func (r *Request) JWTClaim() interface{} {
	return r.HTTPRequest().Context().Value(jwt.CLAIM)
}

// Method returns the HTTP method in uppercase.
func (r *Request) Method() string {
	return strings.ToUpper(r.HTTPRequest().Method)
}

// Is compares the given method with the request HTTP method.
func (r *Request) Is(m string) bool {
	if strings.ToUpper(m) == r.Method() {
		return true
	}
	return false
}

// IsSecure checks if it is a HTTPS request.
func (r *Request) IsSecure() bool {
	return r.Scheme() == "https"
}

// IsPost checks if it is a HTTP POST method.
func (r *Request) IsPost() bool {
	return r.Is(http.MethodPost)
}

// IsGet checks if its a HTTP GET method.
func (r *Request) IsGet() bool {
	return r.Is(http.MethodGet)
}

// IsPatch checks if its a HTTP PATCH method.
func (r *Request) IsPatch() bool {
	return r.Is(http.MethodPatch)
}

// IsPut checks if its a HTTP PUT method.
func (r *Request) IsPut() bool {
	return r.Is(http.MethodPut)
}

// IsDelete checks if its a HTTP DELETE method.
func (r *Request) IsDelete() bool {
	return r.Is(http.MethodDelete)
}

// File returns the file by key.
// It returns a []*FileHeader because the underlying input field could be an array.
// Error will return on parse error or if the key does not exist.
func (r *Request) File(k string) ([]*multipart.FileHeader, error) {
	err := r.parse()
	if err != nil {
		return nil, err
	}

	if val, ok := r.files[k]; ok {
		return val, nil
	}
	return nil, fmt.Errorf(ErrParam, k)
}

// Files returns all existing files.
// It returns a map[string][]*FileHeader because the underlying input field could be an array.
// Error will return on parse error.
func (r *Request) Files() (map[string][]*multipart.FileHeader, error) {
	err := r.parse()
	if err != nil {
		return map[string][]*multipart.FileHeader{}, err
	}
	return r.files, nil
}

// Param returns a parameter by key.
// It returns a []string because the underlying HTML input field could be an array.
// Error will return on parse error or if the key does not exist.
func (r *Request) Param(k string) ([]string, error) {
	err := r.parse()
	if err != nil {
		return nil, err
	}

	if val, ok := r.params[k]; ok {
		return val, nil
	}
	return nil, fmt.Errorf(ErrParam, k)
}

// Params returns all existing parameters.
// It returns a map[string][]string because the underlying HTML input field could be an array.
// Error will return on parse error.
func (r *Request) Params() (map[string][]string, error) {
	err := r.parse()
	if err != nil {
		return nil, err
	}
	return r.params, nil
}

// IP of the request.
func (r *Request) IP() string {
	ips := r.Proxy()
	if len(ips) > 0 && ips[0] != "" {
		rip, _, err := net.SplitHostPort(ips[0])
		if err != nil {
			rip = ips[0]
		}
		return rip
	}
	if ip, _, err := net.SplitHostPort(r.HTTPRequest().RemoteAddr); err == nil {
		return ip
	}
	return r.HTTPRequest().RemoteAddr
}

// Proxy return all IPs which are in the X-Forwarded-For header.
func (r *Request) Proxy() []string {
	if ips := r.HTTPRequest().Header.Get("X-Forwarded-For"); ips != "" {
		return strings.Split(ips, ",")
	}
	return []string{}
}

// Scheme (http/https) checks the `X-Forwarded-Proto` header.
// If that one is empty the URL.Scheme gets checked.
// If that is also empty the request TLS will be checked.
func (r *Request) Scheme() string {
	if scheme := r.HTTPRequest().Header.Get("X-Forwarded-Proto"); scheme != "" {
		return scheme
	}
	if r.HTTPRequest().URL != nil && r.HTTPRequest().URL.Scheme != "" {
		return r.HTTPRequest().URL.Scheme
	}
	if r.HTTPRequest().TLS == nil {
		return "http"
	}
	return "https"
}

// Host returns the host name.
// Port number will be removed if existing.
// If no host info is available, localhost will return.
//		Example: https://example.com:8080/user?id=12#test
//		example.com
func (r *Request) Host() string {
	if r.HTTPRequest().Host != "" {
		if hostPart, _, err := net.SplitHostPort(r.HTTPRequest().Host); err == nil {
			return hostPart
		}
		return r.HTTPRequest().Host
	}
	return "localhost"
}

// Protocol returns the protocol name, such as HTTP/1.1 .
func (r *Request) Protocol() string {
	return r.HTTPRequest().Proto
}

// URI returns full request url with query string, fragment.
//		Example: https://example.com:8080/user?id=12#test
//		/user?id=12#test
func (r *Request) URI() string {
	return r.HTTPRequest().RequestURI
}

// URL returns request url path without the query string and fragment.
//		Example: https://example.com:8080/user?id=12#test
//		/user
func (r *Request) URL() string {
	return r.HTTPRequest().URL.Path
}

// FullURL returns the schema,host,port,uri
//		Example: https://example.com:8080/user?id=12#test
//		https://example.com:8080/user?id=12#test
func (r *Request) FullURL() string {
	s := r.Site()
	if r.Port() != 80 {
		s = fmt.Sprintf("%v:%v%v", s, r.Port(), r.URI())
	}
	return s
}

// Site returns base site url as scheme://domain type without the port.
//		Example: https://example.com:8080/user?id=12#test
//		https://example.com
func (r *Request) Site() string {
	return r.Scheme() + "://" + r.Domain()
}

// Domain is an alias of Host method.
//		Example: https://example.com:8080/user?id=12#test
//		example.com
func (r *Request) Domain() string {
	return r.Host()
}

// Port will return.
// If empty, 80 will be set as default.
func (r *Request) Port() int {
	if _, portPart, err := net.SplitHostPort(r.HTTPRequest().Host); err == nil {
		port, _ := strconv.Atoi(portPart)
		return port
	}
	return 80
}

// Referer returns the Referer Header.
func (r *Request) Referer() string {
	return r.HTTPRequest().Referer()
}

// newRequest is a helper to create a request with the localizer.
func newRequest(r *http.Request) *Request {
	req := &Request{r: r}

	if DefaultLang != "" {
		lang := r.Header.Get("Accept-Language")
		if lang == "" {
			lang = DefaultLang
		}
		req.locale = translation.Localizer(lang)
	}

	return req
}

// parse all router params, get params and post params of a request.
// It runs only once.
// TODO set body limit.
// TODO set filesize limit.
func (r *Request) parse() error {

	if r.params == nil {
		r.params = make(map[string][]string)
		r.files = make(map[string][]*multipart.FileHeader)
	} else {
		//already parsed
		return nil
	}

	// adding router params
	if params := r.HTTPRequest().Context().Value("router_params"); params.(map[string][]string) != nil { // used string instead of router.PATTERN because of dependency cycle.
		r.params = params.(map[string][]string)
	}

	// Handling GET Params
	if r.IsGet() || r.IsDelete() {
		getParams := r.HTTPRequest().URL.Query()
		for param, val := range getParams {
			r.params[param] = val
		}
	}

	// Handling Form Post Params
	if r.IsPost() || r.IsPut() || r.IsPatch() {
		if strings.HasPrefix(r.HTTPRequest().Header.Get("Content-Type"), "multipart/form-data") {
			if err := r.HTTPRequest().ParseMultipartForm(16 * 1024 * 1024); err != nil { //TODO make this customizeable 16MB
				return err
			}
			for file, val := range r.HTTPRequest().MultipartForm.File {
				r.files[file] = val
			}
			for param, val := range r.HTTPRequest().MultipartForm.Value {
				r.params[param] = val
			}
		} else {
			if err := r.HTTPRequest().ParseForm(); err != nil {
				return err
			}
			getParams := r.HTTPRequest().PostForm
			for param, val := range getParams {
				r.params[param] = val
			}
		}
	}

	return nil
}
