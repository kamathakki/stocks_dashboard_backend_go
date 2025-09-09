package registrar

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"stock_automation_backend_go/helper"
	"strings"
	"time"
)

type RouteConfig struct {
	Path          string
	ExactMatch    bool
	Target        string
	Protected     bool
	RewritePrefix string
	Handler       http.Handler
}

type captureResponseWriter struct {
	http.ResponseWriter
	buf    *bytes.Buffer
	status int
}

func (rw *captureResponseWriter) WriteHeader(code int) {
	h := rw.ResponseWriter.Header()
	for _, name := range []string{"Content-Length", "Transfer-Encoding", "Content-Encoding", "Trailer"} {
		h.Del(name)
	}
	rw.status = code
}

func (rw *captureResponseWriter) Write(data []byte) (int, error) {
	return rw.buf.Write(data)
}

func responseWrapperProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rw := &captureResponseWriter{ResponseWriter: w, buf: &bytes.Buffer{}, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		if rw.status >= 400 {
			// fmt.Printf("Captured body: %q\n", rw.buf.String())
			helper.WriteJson(w, rw.status, nil, errors.New(rw.buf.String()))
			return
		}

		helper.WriteJson(w, rw.status, json.RawMessage(rw.buf.Bytes()), nil)
	})
}

func newProxy(base string, stripPrefix string) http.Handler {
	target, _ := url.Parse(base)
	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Transport = &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ResponseHeaderTimeout: 15 * time.Second,
		IdleConnTimeout:       60 * time.Second,
	}

	originalDirector := proxy.Director

	proxy.Director = func(r *http.Request) {
		originalDirector(r)

		r.Host = target.Host
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host
		r.Header.Add("fromgateway", "y")

		if stripPrefix != "" && strings.HasPrefix(r.URL.Path, stripPrefix) {
			r.URL.Path = strings.TrimPrefix(r.URL.Path, stripPrefix)
			if r.URL.Path == "" {
				r.URL.Path = "/"
			}
		}
	}

	return proxy
}

func protectMiddleware(protected bool, mw func(http.Handler) http.Handler, h http.Handler) http.Handler {
	if protected {
		return mw(h)
	}
	return h
}

func ProxyHandler(target string, rewritePrefix string) http.Handler {
	return responseWrapperProxy(newProxy(target, rewritePrefix))
}

func RegisterGatewayRoutes(mux *http.ServeMux, mw func(http.Handler) http.Handler, routes []RouteConfig) {
	for _, route := range routes {
		var h http.Handler
		if route.Handler != nil {
			h = route.Handler
		} else {
			h = ProxyHandler(route.Target, route.RewritePrefix)
		}
		pattern := route.Path
		if !route.ExactMatch {
			if !strings.HasSuffix(pattern, "/") {
				pattern += "/"
			}
		}
		mux.Handle(pattern, protectMiddleware(route.Protected, mw, h))
	}
}
