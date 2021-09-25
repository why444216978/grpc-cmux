package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "helloworld/helloworld"
	"helloworld/server"

	"google.golang.org/grpc"
)

const (
	endpoint = ":8888"
)

type Server struct {
	*server.Server
	pb.UnimplementedGreeterServer
}

func (s *Server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{Message: in.Name + " world"}, nil
}

func registerHTTP(ctx context.Context, s *server.Server) {
	if err := pb.RegisterGreeterHandler(ctx, s.ServerMux, s.GRPClientConn); err != nil {
		panic(err)
	}
}

func registerGRPC(ctx context.Context, s *server.Server) {
	grpcServer := grpc.NewServer()
	pb.RegisterGreeterServer(grpcServer, new(Server))
	if err := grpcServer.Serve(s.GRPCListener); err != nil {
		panic(err)
	}
}

func main() {
	conn, err := net.Listen("tcp", endpoint)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	s := server.New(
		server.WithEndpoint(endpoint),
		server.WithListener(conn),
		server.WithGRPCregisterFunc(registerGRPC),
		server.WithHTTPregisterFunc(registerHTTP))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		panic(s.Start())
		wg.Done()
	}()

	cc, err := newClientConn(endpoint)
	if err != nil {
		log.Fatal(err)
	}

	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		client(cc)
	}

	wg.Wait()
}

func client(cc *grpc.ClientConn) {
	client := pb.NewGreeterClient(cc)

	reply, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "why"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)
}

func newClientConn(target string) (*grpc.ClientConn, error) {
	cc, err := grpc.Dial(
		target,
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	return cc, nil
}
