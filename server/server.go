package main

import (
	"flag"
	"fmt"
	"google.golang.org/grpc/credentials"
	"grpc-practice/hello"
	"io"
	"log"
	"net"
	"strings"

	"google.golang.org/grpc"
)

var (
	tls      = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	certFile = flag.String("cert_file", "", "The TLS cert file")
	keyFile  = flag.String("key_file", "", "The TLS key file")
	port     = flag.Int("port", 10000, "The server port")
	version  = flag.Bool("v", false, "show vesion")
)

var localip string

const ver = "3"

func main() {
	var err error
	if localip, err = localIp(); err != nil {
		localip = "unknown"
	}

	flag.Parse()
	if *version {
		fmt.Println(ver)
		return
	}
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
	err := stream.Send(&hello.HelloResponse{Reply: "er? (from " + localip + ")", Number: []int32{1}})
	if err != nil {
		return err
	}
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		log.Println("recv:", req.Greeting)
		if err := stream.Send(&hello.HelloResponse{Reply: "welcome (from " + localip + ")", Number: []int32{1}}); err != nil {
			return err
		}
	}
}

func localIp() (string, error) {

	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	candidate := ""
	for _, i := range interfaces {
		addresses, err := i.Addrs()
		if err != nil {
			return "", err
		}

		for _, v := range addresses {
			addr := v.String()
			ip := net.ParseIP(addr[:strings.Index(addr, "/")])
			if ip.To4() != nil {
				if strings.HasPrefix(ip.String(), "172") ||
					strings.HasPrefix(ip.String(), "192") ||
					(strings.HasPrefix(ip.String(), "127") && ip.String() != "127.0.0.1") {
					return ip.String(), nil
				} else {
					if candidate == "" {
						candidate = ip.String()
					}
				}
			}
		}
	}
	return candidate, nil
}
