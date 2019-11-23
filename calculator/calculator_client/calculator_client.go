package main

import (
	"context"
	"fmt"
	"github.com/sorborail/grpc-go/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func main() {
	fmt.Println("Client begin start...")

	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect to server: %v", err)
	}
	defer conn.Close()
	cl := calculatorpb.NewCalculatorServiceClient(conn)
	//fmt.Printf("Created client: %f", cl)
	//doUnary(cl)
	//doServerStreaming(cl)
	//doClientStreaming(cl)
	//doNumberStreaming(cl)
	doSquareRoot(cl)
}

func doUnary(cl calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do unary RPC...")
	req := &calculatorpb.SumRequest{
		Values: &calculatorpb.Values{
			FirstValue: 3,
			SecondValue: 10,
		},
	}
	res, err := cl.Calculate(context.Background(), req)
	if err != nil {
		log.Fatalf("Error call rpc request: %v", err)
	}
	fmt.Printf("Response from server: %v", res.Result)
}

func doServerStreaming(cl calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do server streaming RPC...")
	req := &calculatorpb.NumberDecompositionRequest{Value: 120}
	resStream, err := cl.CalculateNumberDecomposition(context.Background(), req)
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
		fmt.Printf("Response from stream server: %v\n", res.GetValue())
	}
}

func doClientStreaming(cl calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do client streaming RPC...")
	stream, err := cl.AverageCalc(context.Background())
	if err != nil {
		log.Fatalf("Error while calling AverageCalc: %v", err)
	}

	var values = [...]int32{2, 6, 78, 90, 232}
	for _, value := range values {
		req := &calculatorpb.AverageCalcRequest{Value: value}
		fmt.Printf("Sending request: %v\n", req)
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while reciving from AverageCalc: %v", err)
	}
	fmt.Printf("AverageCalc response: %v\n", res.GetValue())
}

func doNumberStreaming(cl calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do BiDi number streaming RPC...")

	var tmp int32
	fmt.Println("tmp variable default value: ", tmp)
	stream, err := cl.NumberEveryone(context.Background())
	if err != nil {
		statusCode, _ := status.FromError(err)
		log.Fatalf("Error while calling NumberEveryone: %v", statusCode)
	}
	var requests = [...]calculatorpb.NumberEveryoneRequest{
		calculatorpb.NumberEveryoneRequest{Number: 1},
		calculatorpb.NumberEveryoneRequest{Number: 5},
		calculatorpb.NumberEveryoneRequest{Number: 3},
		calculatorpb.NumberEveryoneRequest{Number: 6},
		calculatorpb.NumberEveryoneRequest{Number: 2},
		calculatorpb.NumberEveryoneRequest{Number: 20},
	}
	wait_ch := make(chan struct{})
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
		close(wait_ch)
	}()
	// block until everything is done
	<-wait_ch
}

func doSquareRoot(cl calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do SquareRoot RPC...")
	var numbers = [...]int32{-10, 18, -24, 38}
	for _, val := range numbers {
		doSquare(cl, val)
	}
}

func doSquare(cl calculatorpb.CalculatorServiceClient, num int32) {
	req := &calculatorpb.SquareRootRequest{Number: num}
	res, err := cl.SquareRoot(context.Background(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			// actual error from gRPC
			fmt.Println(st.Message())
			fmt.Println(st.Code())
			if st.Code() == codes.InvalidArgument {
				fmt.Printf("We probably sent a negative number! %v\n", num)
			}
		} else {
			log.Fatalf("Big error call rpc request: %v", err)
		}

	} else {fmt.Printf("Result of square root number %v: %v\n", num, res.GetNumberRoot())}
}