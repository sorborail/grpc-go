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
	"net"
	"os"
	"strconv"
	"time"
)

type server struct {}

func (*server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Printf("Greet function invoked with request: %v\n", req)
	firstName := req.GetGreeting().GetFirstName()
	res := &greetpb.GreetResponse{
		Result:               "Hello, " + firstName,
	}
	return res, nil
}

func (*server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	fmt.Printf("GreetManyTime function invoked with request: %v\n", req)
	firstName := req.GetGreeting().GetFirstName()
	lastName := req.GetGreeting().GetLastName()
	for i := 0; i < 10; i++ {
		res := &greetpb.GreetManyTimesResponse{Result: "Hello, " + firstName + " " + lastName + "! Number: " + strconv.Itoa(i)}
		err := stream.Send(res)
		if err != nil {
			log.Fatalf("Failed to send response: %v", err)
			return err
		}
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	fmt.Print("LongGreet function invoked with request stream...\n")
	var result string
	i := 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&greetpb.LongGreetResponse{Result: result})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}
		firstName := req.GetGreeting().GetFirstName()
		lastName := req.GetGreeting().GetLastName()
		result += "Hello, " + firstName + " " + lastName + "!"
		i++
		fmt.Println(result + " " + strconv.Itoa(i))
	}
	return nil
}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	fmt.Print("GreetEveryone function invoked with request stream...\n")
	var result string
	i := 0
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}
		firstName := req.GetGreeting().GetFirstName()
		lastName := req.GetGreeting().GetLastName()
		result += "Hello, " + firstName + " " + lastName + "! "
		i++
		fmt.Println(result + " " + strconv.Itoa(i))
		err = stream.Send(&greetpb.GreetEveryoneResponse{Result: result})
		if err != nil {
			log.Fatalf("Error while writing to client stream: %v", err)
		}
	}
	return nil
}

func (*server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {
	fmt.Printf("Greet with deadline function invoked with request: %v\n", req)
	for i := 0; i < 4; i++ {
		if ctx.Err() == context.Canceled {
			fmt.Println("Client canceled request!")
			return nil, status.Error(codes.Canceled, "The client canceled request!")
		}
		time.Sleep(time.Second * 1)
	}
	firstName := req.GetGreeting().GetFirstName()
	lastName := req.GetGreeting().GetLastName()
	res := &greetpb.GreetWithDeadlineResponse{
		Result: "Hello, " + firstName + " " + lastName + "!",
			}
	return res, nil
}

func main() {
	fmt.Println("Hello world!")
	var opts []grpc.ServerOption
	if len(os.Args) > 1 && os.Args[1] == "TLS" {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
		if sslErr != nil {
			log.Fatalf("Error loading server certificate: %v", sslErr)
		}
		opts = append(opts, grpc.Creds(creds))
		fmt.Println("Server started in TLS mode...")
	}
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	srv := grpc.NewServer(opts...)
	greetpb.RegisterGreetServiceServer(srv, &server{})
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("Failed to serve server: %v", err)
	}
}
