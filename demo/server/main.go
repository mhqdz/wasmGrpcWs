package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"nhooyr.io/websocket"

	wasmws "github.com/mhqdz/wasmGrpcWs"
)

//go:generate ./build.bash

type helloServer struct {
	pb.UnimplementedGreeterServer
}

func (*helloServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Println("get request by:", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	fmt.Println("server start")
	//App context setup
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	//Setup HTTP / Websocket server
	router := http.NewServeMux()
	wsl := wasmws.NewWebSocketListener(appCtx, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"}, // 允许所有跨域请求
	})
	router.HandleFunc("/grpc-proxy", wsl.ServeHTTP)
	router.Handle("/", http.FileServer(http.Dir("./static")))
	httpServer := &http.Server{Addr: ":9090", Handler: router}
	//Run HTTP server
	go func() {
		defer appCancel()
		log.Printf("ERROR: HTTP Listen and Server failed; Details: %s", httpServer.ListenAndServe())
	}()

	grpcServer := grpc.NewServer()
	pb.RegisterGreeterServer(grpcServer, new(helloServer))
	//Run gRPC server
	go func() {
		defer appCancel()

		if err := grpcServer.Serve(wsl); err != nil {
			log.Printf("ERROR: Failed to serve gRPC connections; Details: %s", err)
		}
	}()

	//Handle signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		log.Printf("INFO: Received shutdown signal: %s", <-sigs)
		appCancel()
	}()

	//Shutdown
	<-appCtx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*2)
	defer shutdownCancel()

	grpcShutdown := make(chan struct{}, 1)
	go func() {
		grpcServer.GracefulStop()
		grpcShutdown <- struct{}{}
	}()

	httpServer.Shutdown(shutdownCtx)
	select {
	case <-grpcShutdown:
	case <-shutdownCtx.Done():
		grpcServer.Stop()
	}
}
