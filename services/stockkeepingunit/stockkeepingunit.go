package main

import (
	"fmt"
	"net"
	"net/http"
	"stockkeepingunit/env"
	"stockkeepingunit/routes"

	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

type StockKeepingUnitServer struct {
	
}

func main() {
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", env.GetEnv((env.EnvKeys.SKU_PORT))))
	if err != nil {
		fmt.Printf("Error listening to port %v", err)
		return
	}
    m := cmux.New(listener)

	grpcL := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	
	grpcServer := grpc.NewServer()

	httpL := m.Match(cmux.HTTP1Fast())

	server := &http.Server{
		// Addr:    fmt.Sprintf(":%v", env.GetEnv(env.EnvKeys.SKU_PORT)),
		Handler: mux,
	}

	go func() { grpcServer.Serve(grpcL) }()
	go func() { server.Serve(httpL) }()

	fmt.Printf("SKU server and GRPC server is running on port %v. \n", env.GetEnv(env.EnvKeys.SKU_PORT))

	if err := m.Serve(); err != nil {
		fmt.Printf("HTTP server error %v", err)
	}
}
