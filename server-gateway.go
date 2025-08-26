package main

import (
	"fmt"
	"net/http"
	"stock_automation_backend_go/helper"
	"stock_automation_backend_go/shared/env"
	"stock_automation_backend_go/shared/routes"
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
	helper.WriteJson(w, http.StatusOK, res)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", slash)
	mux.HandleFunc("/health", health)
	routes.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", env.GetEnv(env.EnvKeys.BACKEND_PORT)),
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("HTTP server error %v", err)
	}
}
