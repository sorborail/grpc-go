package main

import (
	"context"
	"fmt"
	"github.com/sorborail/grpc-go/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"math"
	"net"
	"strconv"
)

type server struct {}

func (ser *server) Calculate(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	fmt.Printf("Calculate function was invoked with request: %v\n", req)
	res := &calculatorpb.SumResponse{Result: req.GetValues().GetFirstValue() + req.GetValues().GetSecondValue()}
	return res, nil
}

func (*server) CalculateNumberDecomposition(req *calculatorpb.NumberDecompositionRequest,
	stream calculatorpb.CalculatorService_CalculateNumberDecompositionServer) error {
	fmt.Printf("CalculateNumberDecomposition function invoked with request: %v\n", req)
	value := req.GetValue()
	var k int32 = 2
	for value > 1 {
		if value % k == 0 {
			err := stream.Send(&calculatorpb.NumberDecompositionResponse{Value: k})
			if err != nil {
				log.Fatalf("Failed to send response: %v", err)
				return err
			}
			value /= k
		} else {k += 1}
	}
	return nil
}

func (*server) AverageCalc(stream calculatorpb.CalculatorService_AverageCalcServer) error {
	fmt.Print("AverageCalc function invoked with request stream...\n")
	var count, result float64
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&calculatorpb.AverageCalcResponse{Value: result / count})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}
		value := req.GetValue()
		fmt.Println("Current value is: " + strconv.Itoa(int(value)))
		result += float64(value)
		count++
		fmt.Println("Current count is: " + strconv.Itoa(int(count)))
	}
	return nil
}

func (*server) NumberEveryone(stream calculatorpb.CalculatorService_NumberEveryoneServer) error {
	fmt.Println("NumberEveryone function invoked with request stream...")
	var result int32
	ch := make(chan int32)
	defer close(ch)
	//go sendResultToClient(ch, stream)

	go func() {
		for {
			num, opened := <- ch     // получаем данные из потока
			if !opened {
				break    // если поток закрыт, выход из цикла
			}
			err := stream.Send(&calculatorpb.NumberEveryoneResponse{Result: num})
			fmt.Println("Send response number:", num)
			if err != nil {
				log.Fatalf("Error while writing to client stream: %v", err)
			}
		}
		fmt.Println("Server goroutine closed!")
	}()

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}
		num := req.GetNumber()
		fmt.Println("Received number:", num)
		if num > result {
			result = num
			ch <- result
		}
	}
	return nil
}

func sendResultToClient(ch <-chan int32, stream calculatorpb.CalculatorService_NumberEveryoneServer) {
	for {
		switch ch {
		case ch:
			err := stream.Send(&calculatorpb.NumberEveryoneResponse{Result: <-ch})
			if err != nil {
				log.Fatalf("Error while writing to client stream: %v", err)
			}
		}
	}
}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	fmt.Printf("SquareRoot function was invoked with request: %v\n", req)
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Received a negative number! %v", number))
	}
	res := &calculatorpb.SquareRootResponse{NumberRoot: math.Sqrt(float64(number))}
	return res, nil
}

func main() {
	fmt.Println("Calculator server start...")

	lst, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("Failed start listen network on adderess localhost:50051: %v", err)
	}

	srv := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(srv, &server{})

	// Register reflection service on gRPC server
	reflection.Register(srv)
	if err := srv.Serve(lst); err != nil {
		log.Fatalf("Failed to serve server: %v", err)
	}
}
