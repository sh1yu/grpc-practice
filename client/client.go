package main

import (
	"context"
	"flag"
	"github.com/sercand/kuberesolver/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"grpc-practice/hello"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	tls           = flag.Bool("tls", false, "Connection uses TLS if true, else plain TCP")
	caFile        = flag.String("crt", "", "The file containing the CA root cert file")
	serverAddr    = flag.String("addr", "psy.cn:443", "The server address in the format of host:port")
	crtServerName = flag.String("host", "psy.cn", "The server name use to verify the hostname returned by TLS handshake")
	c             = flag.Int("c", 5, "concurrent")
	n             = flag.Int("n", 10000, "request for each concurrent")
	s             = flag.Int("s", 500, "sleep milliseconds between each request")
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "canary: always"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var headerFlags arrayFlags

func main() {
	flag.Var(&headerFlags, "H", "grpc request header")
	flag.Parse()
	log.SetOutput(os.Stdout)

	headerKvs := make([]string, 0, 2*len(headerFlags))
	for _, v := range headerFlags {
		token := strings.Split(v, ":")
		if len(token) != 2 {
			continue
		}
		headerKvs = append(headerKvs, strings.TrimSpace(token[0]), strings.TrimSpace(token[1]))
	}

	var opts []grpc.DialOption
	if *tls {
		//with tls
		if *caFile == "" {
			*caFile = "crt/cert.pem"
		}
		creds, err := credentials.NewClientTLSFromFile(*caFile, *crtServerName)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		//without tls
		opts = append(opts, grpc.WithInsecure())
	}

	opts = append(opts, grpc.WithBlock())

	//with manual scheme
	//rb := manual.NewBuilderWithScheme("psy")
	//rb.InitialState(resolver.State{Addresses: []resolver.Address{{Addr: "localhost:10000"}}})
	//resolver.Register(rb)

	//ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	//conn, err := grpc.DialContext(ctx, "psy:///", opts...)

	//with passthrough scheme
	//ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	//conn, err := grpc.DialContext(ctx, "passthrough:///"+*serverAddr, opts...)

	//with k8s scheme
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
	kuberesolver.RegisterInCluster()
	conn, err := grpc.DialContext(ctx, "kubernetes:///service.namespace:portname", opts...)

	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	defer func() {
		_ = conn.Close()
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < *c; i++ {
		wg.Add(1)
		go func(index int) {
			for i := 0; i < *n; i++ {
				doRequest(conn, headerKvs, index)
				<-time.After(time.Duration(*s) * time.Millisecond)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func doRequest(conn *grpc.ClientConn, headerKvs []string, index int) {
	client := hello.NewHelloServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	metaCtx := metadata.AppendToOutgoingContext(ctx, headerKvs...)
	stream, err := client.SayHello(metaCtx)
	if err != nil {
		log.Fatalf("[%v]  %v.SayHello, %v", index, client, err)
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
				log.Printf("[%v]  Failed to receive a reply : %v", index, err)
				close(waitc)
				return
			}
			log.Printf("[%v] Got message %s ", index, in.Reply)
		}
	}()
	for i := 0; i < 1; i++ {
		if err := stream.Send(&hello.HelloRequest{Greeting: "good morning! " + strconv.Itoa(i)}); err != nil {
			log.Fatalf("Failed to send a request: %v", err)
		}
		//<-time.After(1 * time.Millisecond)
	}
	_ = stream.CloseSend()
	<-waitc
}
