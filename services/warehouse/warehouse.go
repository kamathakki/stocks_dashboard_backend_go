package main

import (
	"fmt"
	"net/http"
	"warehouse/env"
	"warehouse/routes"
)

func main() {
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", env.GetEnv(env.EnvKeys.WAREHOUSE_PORT)),
		Handler: mux,
	}

	fmt.Printf("Warehouse server is running on port %v. \n", env.GetEnv(env.EnvKeys.WAREHOUSE_PORT))

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("HTTP server error %v", err)
	}
}
