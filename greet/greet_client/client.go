package main

import (
	"context"
	"fmt"
	"github.com/sorborail/grpc-go/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"time"
)

func main() {
	fmt.Println("Hello, I'm a client!")
	opts := grpc.WithInsecure()
	if len(os.Args) > 1 && os.Args[1] == "TLS" {
		certFile := "ssl/ca.crt" //Certificate authority trust certificate
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
		if sslErr != nil {
			log.Fatalf("Error while loading CA trust certificate %v", sslErr)
		}
		opts = grpc.WithTransportCredentials(creds)
		fmt.Println("Client started in TLS mode...")
	}
	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("Could not connect to server: %v", err)
	}
	defer conn.Close()
	cl := greetpb.NewGreetServiceClient(conn)
	//fmt.Printf("Created client: %f", cl)
	doUnary(cl)
	doServerStreaming(cl)
	doClientStreaming(cl)
	doBiDiStreaming(cl)
	doUnaryWithDeadline(cl, 5 * time.Second) //should complete
	doUnaryWithDeadline(cl, 1 * time.Second) //should timeout
}

func doUnary(cl greetpb.GreetServiceClient) {
	fmt.Println("Starting to do unary RPC...")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName:            "Alexey",
			LastName:             "Kharitonov",
		},
	}
	res, err := cl.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Error call rpc request: %v", err)
	}
	fmt.Printf("Response from server: %v", res.Result)
}

func doServerStreaming(cl greetpb.GreetServiceClient) {
	fmt.Println("Starting to do server streaming RPC...")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName:            "Alexey",
			LastName:             "Kharitonov",
		},
	}
	resStream, err := cl.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error call server stream rpc request: %v", err)
	}
	for {
		res, err := resStream.Recv()
		// EOF - close stream
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed reading stream from server: %v", err)
		}
		fmt.Printf("Response from stream server: %v\n", res.GetResult())
	}
}

func doClientStreaming(cl greetpb.GreetServiceClient) {
	fmt.Println("Starting to do client streaming RPC...")
	stream, err := cl.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet: %v", err)
	}
	req := &greetpb.LongGreetRequest{Greeting: &greetpb.Greeting{FirstName: "Alexey", LastName: "Kharitonov"}}
	for i := 0; i < 5; i++ {
		fmt.Printf("Sending request: %v\n", req)
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while reciving from LongGreet: %v", err)
	}
	fmt.Printf("LongGreet response: %v\n", res)
}

func doBiDiStreaming(cl greetpb.GreetServiceClient) {
	fmt.Println("Starting to do BiDi streaming RPC...")

	// we create a stream by invoking the client
	stream, err := cl.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while calling GreetEveryone: %v", err)
	}
	var requests = [...]greetpb.GreetEveryoneRequest{
		greetpb.GreetEveryoneRequest{Greeting: &greetpb.Greeting{FirstName: "Ivan", LastName: "Ivanov"}},
		greetpb.GreetEveryoneRequest{Greeting: &greetpb.Greeting{FirstName: "Petr", LastName: "Petrov"}},
		greetpb.GreetEveryoneRequest{Greeting: &greetpb.Greeting{FirstName: "Semen", LastName: "Semenov"}},
		greetpb.GreetEveryoneRequest{Greeting: &greetpb.Greeting{FirstName: "Bob", LastName: "Marcy"}},
		greetpb.GreetEveryoneRequest{Greeting: &greetpb.Greeting{FirstName: "Den", LastName: "Symon"}},
	}
	waitch := make(chan struct{})
	// we send a bunch of messages to the client (go routine)
	go func() {
		// to send a bunch of messages
		for _, req := range requests {
			fmt.Printf("Sending request: %v\n", req)
			_ = stream.Send(&req)
			time.Sleep(1000 * time.Millisecond)
		}
		_ = stream.CloseSend()
	}()
	// we receive a bunch of messages from the client (go routine)
	go func() {
		// to receive a bunch of messages
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while reciving from GreetAveryone: %v", err)
			}
			fmt.Printf("Received response: %v\n", res.GetResult())
		}
		close(waitch)
	}()
	// block until everything is done
	<-waitch
}

func doUnaryWithDeadline(cl greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Println("Starting to do unary with deadline RPC...")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName:            "Alexey",
			LastName:             "Kharitonov",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer  cancel()
	res, err := cl.GreetWithDeadline(ctx, req)
	if err != nil {
		stErr, ok := status.FromError(err)
		if ok {
			if stErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout with hit! Deadline was exeeded!")
			} else {
				fmt.Printf("Unexpected error! %v\n", stErr)
			}
		} else {
			log.Fatalf("Error call rpc GreetWithDeadline request: %v", err)
		}
	} else {
		fmt.Printf("Response from server greetWithDeadline: %v\n", res.GetResult())
	}
}
