package main

import (
	"fmt"
	routes "iam/routes"
	"net/http"

	"iam/env"
	"stock_automation_backend_go/database/redis"
)

func main() {
	redis.InitRedis()
	fmt.Println("Redis cache connected")
	defer redis.QuitRedis()
	mux := http.NewServeMux()

	routes.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", env.GetEnv[string](env.EnvKeys.IAM_PORT)),
		Handler: mux,
	}

	fmt.Printf("IAM server is running on port %v. \n", env.GetEnv[string](env.EnvKeys.IAM_PORT))

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("HTTP server error %v", err)
	}
}
