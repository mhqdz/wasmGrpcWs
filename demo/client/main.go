package main

import (
	"context"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"

	wasmws "github.com/mhqdz/wasmGrpcWs"
)

func main() {
	//Connect to remote gRPC server
	const websocketURL = "ws://localhost:9090/grpc-proxy"
	// creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
	conn, err := grpc.DialContext(context.Background(), "passthrough:///"+websocketURL, grpc.WithContextDialer(wasmws.GRPCDialer), grpc.WithDisableRetry(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not gRPC dial: %s; Details: %s", websocketURL, err)
	}
	defer conn.Close()

	//Test setup
	client := pb.NewGreeterClient(conn)

	reply, err := client.SayHello(context.Background(), &pb.HelloRequest{
		Name: "mhqdz",
	})

	fmt.Println(reply, err)
}
