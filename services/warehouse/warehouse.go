package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"stock_automation_backend_go/database/redis"
	"warehouse/env"
	stockCountJob "warehouse/job"
	updatestockcountforwarehouselocationpb "warehouse/proto"
	"warehouse/routes"
	"warehouse/warehouseendpoints"

	_ "github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

func main() {
	redis.InitRedis()
	fmt.Println("Redis cache connected")
	defer redis.QuitRedis()
    mux := http.NewServeMux()

    // If WAREHOUSE_GRPC_PORT is provided, run HTTP and gRPC on separate ports.
    if grpcPort, ok := os.LookupEnv("WAREHOUSE_GRPC_PORT"); ok && grpcPort != "" {
        routes.RegisterRoutes(mux)

        httpPort := env.GetEnv[string](env.EnvKeys.WAREHOUSE_PORT)

        grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", grpcPort))
        if err != nil {
            fmt.Printf("Error listening to grpc port %v", err)
            return
        }

        grpcServer := grpc.NewServer()
        updatestockcountforwarehouselocationpb.RegisterWarehouseServer(grpcServer, &warehouseendpoints.WarehouseServer{})

        go stockCountJob.RunJob()
        go func() { grpcServer.Serve(grpcListener) }()
        go func() { http.ListenAndServe(fmt.Sprintf(":%v", httpPort), mux) }()
        fmt.Printf("Warehouse HTTP running on %v, gRPC running on %s. \n", httpPort, grpcPort)
        select {}
    } else {
        log.Fatal("WAREHOUSE_GRPC_PORT is not set")
        // listener, err := net.Listen("tcp", fmt.Sprintf(":%v", env.GetEnv[string](env.EnvKeys.WAREHOUSE_PORT)))
        // if err != nil {
        //     fmt.Printf("Error listening to port %v", err)
        //     return
        // }

        // m := cmux.New(listener)

        // grpcL := m.MatchWithWriters(
        //     cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
        // )
        // grpcServer := grpc.NewServer()

        // updatestockcountforwarehouselocationpb.RegisterWarehouseServer(grpcServer, &warehouseendpoints.WarehouseServer{})

        // httpL := m.Match(cmux.Any())
        
        // routes.RegisterRoutes(mux)

        // server := &http.Server{
        //     Handler: mux,
        // }

        // go stockCountJob.RunJob()
        // go func() { grpcServer.Serve(grpcL) }()
        // go func() { server.Serve(httpL) }()
        // fmt.Printf("Warehouse server and GRPC server is running on port %v. \n", env.GetEnv[string](env.EnvKeys.WAREHOUSE_PORT))

        // if err := m.Serve(); err != nil {
        //     fmt.Printf("Error in configuration %v", err)
        // }
    }
	
}
