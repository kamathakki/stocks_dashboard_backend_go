package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"stock_automation_backend_go/database/redis"
	"stockkeepingunit/env"
	"stockkeepingunit/routes"

	_ "github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

type StockKeepingUnitServer struct {
	
}

func main() {
	redis.InitRedis()
	fmt.Println("Redis cache connected")
	defer redis.QuitRedis()
    mux := http.NewServeMux()
    routes.RegisterRoutes(mux)

    if grpcPort, ok := os.LookupEnv("SKU_GRPC_PORT"); ok && grpcPort != "" {
        httpPort := env.GetEnv[string](env.EnvKeys.SKU_PORT)

        grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
        if err != nil {
            fmt.Printf("Error listening to grpc port %v", err)
            return
        }

        grpcServer := grpc.NewServer()

        go func() { grpcServer.Serve(grpcListener) }()
        go func() { http.ListenAndServe(fmt.Sprintf(":%v", httpPort), mux) }()
        fmt.Printf("SKU HTTP running on %v, gRPC running on %s. \n", httpPort, grpcPort)
        select {}
    } else {
        log.Fatal("SKU_GRPC_PORT is not set")
        // listener, err := net.Listen("tcp", fmt.Sprintf(":%v", env.GetEnv[string]((env.EnvKeys.SKU_PORT))))
        // if err != nil {
        //     fmt.Printf("Error listening to port %v", err)
        //     return
        // }
        // m := cmux.New(listener)

        // grpcL := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
        
        // grpcServer := grpc.NewServer()

        // httpL := m.Match(cmux.HTTP1Fast())

        // server := &http.Server{
        //     Handler: mux,
        // }

        // go func() { grpcServer.Serve(grpcL) }()
        // go func() { server.Serve(httpL) }()

        // fmt.Printf("SKU server and GRPC server is running on port %v. \n", env.GetEnv[string](env.EnvKeys.SKU_PORT))

        // if err := m.Serve(); err != nil {
        //     fmt.Printf("HTTP server error %v", err)
        // }
    }
}
