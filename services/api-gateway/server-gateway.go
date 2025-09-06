package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"stock_automation_backend_go/helper"
	"stock_automation_backend_go/services/socketio"
	"stock_automation_backend_go/shared/env"
	"strings"
	"time"
)

type ResponseStruct struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func slash(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to Stock Automation Backend in GOLang"))
}

func health(w http.ResponseWriter, r *http.Request) {
	res := ResponseStruct{StatusCode: http.StatusOK, Message: "OK"}
	helper.WriteJson(w, http.StatusOK, res, nil)
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

func ResponseWrapperProxy(next http.Handler) http.Handler {
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

func main() {
	socketio.RegisterHandlers()
	socketServer := socketio.GetServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/", slash)
	mux.HandleFunc("/health", health)

	iam := fmt.Sprintf("http://localhost:%s", env.GetEnv[string](env.EnvKeys.IAM_PORT))
	sku := fmt.Sprintf("http://localhost:%s", env.GetEnv[string](env.EnvKeys.SKU_PORT))
	wh := fmt.Sprintf("http://localhost:%s", env.GetEnv[string](env.EnvKeys.WAREHOUSE_PORT))

	mux.Handle("/api/iam/", ResponseWrapperProxy(newProxy(iam, "/api/iam")))
	mux.Handle("/api/warehouse/", ResponseWrapperProxy(newProxy(wh, "/api/warehouse")))
	mux.Handle("/api/stockkeepingunit/", ResponseWrapperProxy(newProxy(sku, "/api/stockkeepingunit")))
	mux.Handle("/socket.io/", socketServer)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", env.GetEnv[string](env.EnvKeys.BACKEND_PORT)),
		Handler: mux,
	}

	go func() {
	//defer socketServer.Close()
	socketServer.Serve()
	defer socketServer.Close()
	}()
	fmt.Println("Socket server connected")

	// redis.InitRedis()
	// fmt.Println("Redis cache connected")
	// defer redis.QuitRedis()

	fmt.Printf("API Gateway is running on port %v. \n", env.GetEnv[string](env.EnvKeys.BACKEND_PORT))

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("HTTP server error %v", err)
	}



}
