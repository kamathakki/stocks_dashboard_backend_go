package main

import (
	"fmt"
	"net"
	"net/http"
	"warehouse/env"
	updatestockcountforwarehouselocationpb "warehouse/proto"
	"warehouse/routes"
	"warehouse/warehouseendpoints"

	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

func main() {
	mux := http.NewServeMux()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%v", env.GetEnv(env.EnvKeys.WAREHOUSE_PORT)))
	if err != nil {
		fmt.Printf("Error listening to port %v", err)
		return
	}

	m := cmux.New(listener)

	grpcL := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	grpcServer := grpc.NewServer()

	updatestockcountforwarehouselocationpb.RegisterWarehouseServer(grpcServer, &warehouseendpoints.WarehouseServer{})

	httpL := m.Match(cmux.HTTP1Fast())
	
	routes.RegisterRoutes(mux)

	server := &http.Server{
		// Addr:    fmt.Sprintf(":%v", env.GetEnv(env.EnvKeys.WAREHOUSE_PORT)),
		Handler: mux,
	}

	go func() { grpcServer.Serve(grpcL) }()
	go func() { server.Serve(httpL) }()

	fmt.Printf("Warehouse server and GRPC server is running on port %v. \n", env.GetEnv(env.EnvKeys.WAREHOUSE_PORT))

	if err := m.Serve(); err != nil {
		fmt.Printf("Error in configuration %v", err)
	}
}
