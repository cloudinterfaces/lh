package lh

import (
	"fmt"
	"net/http"
	"net/textproto"
)

var mh = textproto.CanonicalMIMEHeaderKey

// DemangleInputHeaders controls the attempt
// to replace X-Amzn-Remapped-* headers with
// non-remapped equivalents. If true, the X-Amzn-Remapped
// header is deleted if it can be unremapped.
// If the non-remapped
// header exists and is not empty, the
// X-Amzn-Remapped header will not be deleted.
// For example, the Via header will generally
// be supplied by API Gateway, therefore
// X-Amzn-Remapped-Via may still be present if the
// client sent a Via header.
// This should be set (if needed) prior to calling ServeHTTP.
var DemangleInputHeaders = false

// MangleOutputHeaders controls
// replacing output headers with X-Amzn-Remapped
// headers. If set, headers that need to be remapped
// have an additional X-Amzn-Remapped header created
// with the same value. This should be
// set (if needed) prior to calling ServeHTTP.
var MangleOutputHeaders = false

// Mangle is the set of headers API Gateway
// may prefer as X-Amzn-Remapped headers.
// It's a good practice to reduce this
// to a set of concern prior to calling
// ServeHTTP if MangleOutputHeaders is true.
var Mangle = map[string]struct{}{
	"Accept":             {},
	"Accept-Charset":     {},
	"Accept-Encoding":    {},
	"Age":                {},
	"Authorization":      {},
	"Connection":         {},
	"Content-Encoding":   {},
	"Content-Length":     {},
	"Content-MD5":        {},
	"Content-Type":       {},
	"Date":               {},
	"Expect":             {},
	"Host":               {},
	"Max-Forwards":       {},
	"Pragma":             {},
	"Proxy-Authenticate": {},
	"Range":              {},
	"Referer":            {},
	"Server":             {},
	"TE":                 {},
	"Trailer":            {},
	"Transfer-Encoding":  {},
	"Upgrade":            {},
	"User-Agent":         {},
	"Via":                {},
	"WWW-Authenticate":   {},
	"Warn":               {},
}

// Demangle is the set of possible X-Amzn-Remapped
// headers. It is a good practice to reduce this
// to a set of concern prior to calling ServeHTTP if DemangleInputHeaders
// is true.
var Demangle = map[string]struct{}{
	"X-Amzn-Remapped-Accept":             {},
	"X-Amzn-Remapped-Accept-Charset":     {},
	"X-Amzn-Remapped-Accept-Encoding":    {},
	"X-Amzn-Remapped-Age":                {},
	"X-Amzn-Remapped-Authorization":      {},
	"X-Amzn-Remapped-Connection":         {},
	"X-Amzn-Remapped-Content-Encoding":   {},
	"X-Amzn-Remapped-Content-Length":     {},
	"X-Amzn-Remapped-Content-MD5":        {},
	"X-Amzn-Remapped-Content-Type":       {},
	"X-Amzn-Remapped-Date":               {},
	"X-Amzn-Remapped-Expect":             {},
	"X-Amzn-Remapped-Host":               {},
	"X-Amzn-Remapped-Max-Forwards":       {},
	"X-Amzn-Remapped-Pragma":             {},
	"X-Amzn-Remapped-Proxy-Authenticate": {},
	"X-Amzn-Remapped-Range":              {},
	"X-Amzn-Remapped-Referer":            {},
	"X-Amzn-Remapped-Server":             {},
	"X-Amzn-Remapped-TE":                 {},
	"X-Amzn-Remapped-Trailer":            {},
	"X-Amzn-Remapped-Transfer-Encoding":  {},
	"X-Amzn-Remapped-Upgrade":            {},
	"X-Amzn-Remapped-User-Agent":         {},
	"X-Amzn-Remapped-Via":                {},
	"X-Amzn-Remapped-WWW-Authenticate":   {},
	"X-Amzn-Remapped-Warn":               {},
}

func mangle(h http.Header) {
	for k := range h {
		k = mh(k)
		if _, ok := Mangle[k]; ok {
			newkey := fmt.Sprintf("X-Amzn-Remapped-%s", k)
			if len(h.Get(newkey)) == 0 {
				v := h.Get(k)
				h.Set(newkey, v)
			}
		}
	}
}

func demangle(h http.Header) {
	for k := range Demangle {
		if v := h.Get(k); len(v) > 0 {
			newkey := k[16:]
			if len(h.Get(newkey)) == 0 {
				h.Set(newkey, v)
			}
		}
	}
}
