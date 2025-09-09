package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"stock_automation_backend_go/services/api-gateway/middleware"
	"stock_automation_backend_go/services/api-gateway/middleware/registrar"
	"strings"

	"stock_automation_backend_go/services/socketio"
	"stock_automation_backend_go/shared/env"
)

type ResponseStruct struct {
	StatusCode int  `json:"statusCode"`
	Status     bool `json:"status"`
}

func slash(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to Stock Automation Backend in GOLang"))
}

func health(w http.ResponseWriter, r *http.Request) {

	jsonData, err := json.Marshal(ResponseStruct{StatusCode: http.StatusOK, Status: true})
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// type captureResponseWriter struct {
// 	http.ResponseWriter
// 	buf    *bytes.Buffer
// 	status int
// }

// func (rw *captureResponseWriter) WriteHeader(code int) {
// 	h := rw.ResponseWriter.Header()
// 	for _, name := range []string{"Content-Length", "Transfer-Encoding", "Content-Encoding", "Trailer"} {
// 		h.Del(name)
// 	}
// 	rw.status = code
// }

// func (rw *captureResponseWriter) Write(data []byte) (int, error) {
// 	return rw.buf.Write(data)
// }

// func ResponseWrapperProxy(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		rw := &captureResponseWriter{ResponseWriter: w, buf: &bytes.Buffer{}, status: http.StatusOK}
// 		next.ServeHTTP(rw, r)

// 		if rw.status >= 400 {
// 			// fmt.Printf("Captured body: %q\n", rw.buf.String())
// 			helper.WriteJson(w, rw.status, nil, errors.New(rw.buf.String()))
// 			return
// 		}

// 		helper.WriteJson(w, rw.status, json.RawMessage(rw.buf.Bytes()), nil)
// 	})
// }

func corsMiddleware(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Bypass CORS handling for Socket.IO to avoid interfering with engine.io
			if r.URL.Path == "/socket.io" || strings.HasPrefix(r.URL.Path, "/socket.io/") {
				next.ServeHTTP(w, r)
				return
			}
			if allowedOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Requested-With")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// func newProxy(base string, stripPrefix string) http.Handler {
// 	target, _ := url.Parse(base)
// 	proxy := httputil.NewSingleHostReverseProxy(target)

// 	proxy.Transport = &http.Transport{
// 		Proxy:                 http.ProxyFromEnvironment,
// 		ResponseHeaderTimeout: 15 * time.Second,
// 		IdleConnTimeout:       60 * time.Second,
// 	}

// 	originalDirector := proxy.Director

// 	proxy.Director = func(r *http.Request) {
// 		originalDirector(r)

// 		r.Host = target.Host
// 		r.URL.Scheme = target.Scheme
// 		r.URL.Host = target.Host
// 		r.Header.Add("fromgateway", "y")

// 		if stripPrefix != "" && strings.HasPrefix(r.URL.Path, stripPrefix) {
// 			r.URL.Path = strings.TrimPrefix(r.URL.Path, stripPrefix)
// 			if r.URL.Path == "" {
// 				r.URL.Path = "/"
// 			}
// 		}
// 	}

// 	return proxy
// }

func main() {
	socketServer := socketio.GetServer()

	mux := http.NewServeMux()
	mux.HandleFunc("/", slash)
	mux.HandleFunc("/health", health)

	// Refresh token endpoint mirrors TS verifyRefreshToken
	mux.HandleFunc("/api/iam/refresh", middleware.VerifyRefreshTokenHandler)

	iam := fmt.Sprintf("http://localhost:%s", env.GetEnv[string](env.EnvKeys.IAM_PORT))
	sku := fmt.Sprintf("http://localhost:%s", env.GetEnv[string](env.EnvKeys.SKU_PORT))
	wh := fmt.Sprintf("http://localhost:%s", env.GetEnv[string](env.EnvKeys.WAREHOUSE_PORT))

	mux.Handle("/socket.io/", socketServer)

	routes := []registrar.RouteConfig{
		{
			Path:          "/api/iam/login",
			ExactMatch:    true,
			Target:        iam,
			Protected:     false,
			RewritePrefix: "/api/iam",
			Handler:       registrar.ProxyHandler(iam, "/api/iam"),
		},
		{
			Path:          "/api/iam/register",
			ExactMatch:    true,
			Target:        iam,
			Protected:     false,
			RewritePrefix: "/api/iam",
			Handler:       registrar.ProxyHandler(iam, "/api/iam"),
		},
		{
			Path:          "/api/iam",
			ExactMatch:    false,
			Target:        iam,
			Protected:     true,
			RewritePrefix: "/api/iam",
			Handler:       registrar.ProxyHandler(iam, "/api/iam"),
		},
		{
			Path:          "/api/warehouse",
			ExactMatch:    false,
			Target:        wh,
			Protected:     true,
			RewritePrefix: "/api/warehouse",
			Handler:       registrar.ProxyHandler(wh, "/api/warehouse"),
		},
		{
			Path:          "/api/stockkeepingunit",
			ExactMatch:    false,
			Target:        sku,
			Protected:     true,
			RewritePrefix: "/api/stockkeepingunit",
			Handler:       registrar.ProxyHandler(sku, "/api/stockkeepingunit"),
		},
	}

	registrar.RegisterGatewayRoutes(mux, middleware.VerifyTokenMiddleware, routes)

	// Protect selected routes with VerifyTokenMiddleware similar to TS verifyToken
	// mux.Handle("/api/iam/", ResponseWrapperProxy(newProxy(iam, "/api/iam")))
	// mux.Handle("/api/warehouse/", middleware.VerifyTokenMiddleware(ResponseWrapperProxy(newProxy(wh, "/api/warehouse"))))
	// mux.Handle("/api/stockkeepingunit/", middleware.VerifyTokenMiddleware(ResponseWrapperProxy(newProxy(sku, "/api/stockkeepingunit"))))

	// Build allowed origin from env: FRONTEND_PROTOCOL://FRONTEND_CLIENT
	allowedOrigin := fmt.Sprintf("%s://%s", env.GetEnv[string](env.EnvKeys.FRONTEND_PROTOCOL), env.GetEnv[string](env.EnvKeys.FRONTEND_CLIENT))

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", env.GetEnv[string](env.EnvKeys.BACKEND_PORT)),
		Handler: corsMiddleware(allowedOrigin)(mux),
	}

	go func() {
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
