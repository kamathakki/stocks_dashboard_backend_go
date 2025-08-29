package main

import (
	"fmt"
	"net/http"
	"stock_automation_backend_go/services/stockkeepingunit/routes"
	"stock_automation_backend_go/shared/env"
)

func main() {
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", env.GetEnv(env.EnvKeys.SKU_PORT)),
		Handler: mux,
	}

	fmt.Printf("SKU server is running on port %v. \n", env.GetEnv(env.EnvKeys.SKU_PORT))

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("HTTP server error %v", err)
	}
}
