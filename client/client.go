package main

import (
	"context"
	"flag"
	"google.golang.org/grpc"
	"grpc-practice/hello"
	"io"
	"log"
	"time"
)

var (
	//tls                = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	//caFile             = flag.String("ca_file", "", "The file containing the CA root cert file")
	serverAddr = flag.String("server_addr", "chart-example.local:443", "The server address in the format of host:port")
	//serverHostOverride = flag.String("server_host_override", "x.test.youtube.com", "The server name use to verify the hostname returned by TLS handshake")
)

func main() {
	flag.Parse()
	var opts []grpc.DialOption

	//with tls
	//if *tls {
	//	if *caFile == "" {
	//		*caFile = testdata.Path("ca.pem")
	//	}
	//	creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostOverride)
	//	if err != nil {
	//		log.Fatalf("Failed to create TLS credentials %v", err)
	//	}
	//	opts = append(opts, grpc.WithTransportCredentials(creds))
	//} else {
	//	opts = append(opts, grpc.WithInsecure())
	//}
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	//with manual scheme
	//rb := manual.NewBuilderWithScheme("psy")
	//rb.InitialState(resolver.State{Addresses: []resolver.Address{{Addr: "localhost:10000"}}})
	//resolver.Register(rb)

	//ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	//conn, err := grpc.DialContext(ctx, "psy:///", opts...)

	//with passthrough scheme
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	conn, err := grpc.DialContext(ctx, "passthrough:///"+*serverAddr, opts...)

	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	defer func() {
		_ = conn.Close()
	}()

	client := hello.NewHelloServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := client.SayHello(ctx)
	if err != nil {
		log.Fatalf("%v.SayHello, %v", client, err)
	}
	waitc := make(chan struct{})
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a reply : %v", err)
			}
			log.Printf("Got message %s ", in.Reply)
		}
	}()
	for i := 0; i < 5; i++ {
		if err := stream.Send(&hello.HelloRequest{Greeting: "good morning!"}); err != nil {
			log.Fatalf("Failed to send a request: %v", err)
		}
		<-time.After(1 * time.Second)
	}
	_ = stream.CloseSend()
	<-waitc
}
