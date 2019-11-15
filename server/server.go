package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc/credentials"
	"grpc-practice/hello"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
)

var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "", "The TLS cert file")
	keyFile  = flag.String("key_file", "", "The TLS key file")
	port     = flag.Int("port", 10000, "The server port")
)

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	//with tls
	if *tls {
		if *certFile == "" {
			*certFile = "crt/cert.pem"
		}
		if *keyFile == "" {
			*keyFile = "crt/key.pem"
		}
		creds, err := credentials.NewServerTLSFromFile(*certFile, *keyFile)
		if err != nil {
			log.Fatalf("Failed to generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}

	grpcServer := grpc.NewServer(opts...)

	hello.RegisterHelloServiceServer(grpcServer, &helloServer{})
	log.Printf("grpc listening at port: %v", *port)
	log.Fatalf("%v", grpcServer.Serve(lis))
}

type helloServer struct {
}

func (s *helloServer) SayHello(stream hello.HelloService_SayHelloServer) error {
	err := stream.Send(&hello.HelloResponse{Reply: "er?", Number: []int32{1}})
	if err != nil {
		return err
	}
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.Send(&hello.HelloResponse{Reply: "welcome", Number: []int32{1}})
		}
		if err != nil {
			return err
		}
		log.Println("recv:", req.Greeting)
	}
}
