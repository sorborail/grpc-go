package main

import (
	"context"
	"fmt"
	"github.com/sorborail/grpc-go/blog/blogpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"os"
)

func main() {
	fmt.Println("Starting Blog client...!")
	opts := grpc.WithInsecure()
	if len(os.Args) > 1 && os.Args[1] == "TLS" {
		certFile := "ssl/ca.crt" //Certificate authority trust certificate
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
		if sslErr != nil {
			log.Fatalf("Error while loading CA trust certificate %v", sslErr)
		}
		opts = grpc.WithTransportCredentials(creds)
		fmt.Println("Blog client started in TLS mode...")
	}

	conn, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("Could not connect to Blog server: %v", err)
	}
	defer conn.Close()
	cl := blogpb.NewBlogServiceClient(conn)
	fmt.Println("Blog client is started.")
	//doCreateBlog(cl)
	//doReadBlog(cl)
	//doUpdateBlog(cl)
	//doDeleteBlog(cl)
	doListBlog(cl)
}

// Create Blog function
func doCreateBlog(cl blogpb.BlogServiceClient) {
	fmt.Println("Begin creating blog...")
	req := &blogpb.CreateBlogRequest{
		Blog: &blogpb.Blog{
			AuthorId:             "Tima",
			Title:                "Tima's Blog",
			Content:              "Big content of the my new blog.",
		},
	}
	res, err := cl.CreateBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("Unexpected error: %v\n", err)
	}
	fmt.Printf("Blog has been created: %v\n", res.GetBlog())
}

func doReadBlog(cl blogpb.BlogServiceClient) {
	fmt.Println("Reading the blog...")
	req := &blogpb.ReadBlogRequest{
		BlogId:               "5dd702170004b63fd57da98c",
	}
	res, err := cl.ReadBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("Error happened while reading: %v\n", err)
	}
	fmt.Printf("Blog has been readed: %v\n", res.GetBlog())
}

func doUpdateBlog(cl blogpb.BlogServiceClient) {
	fmt.Println("Begin updating the blog...")
	req := &blogpb.UpdateBlogRequest{
		Blog:               &blogpb.Blog{
			Id:                   "5dd702170004b63fd57da98c",
			AuthorId:             "Alexey",
			Title:                "New blog title for fist blog",
			Content:              "New content for updated first blog!",
		},
	}
	res, err := cl.UpdateBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("Error happened while updating: %v\n", err)
	}
	fmt.Printf("Blog has been updated: %v\n", res.GetBlog())
}

func doDeleteBlog(cl blogpb.BlogServiceClient) {
	fmt.Println("Begin deleting blog...")
	req := &blogpb.DeleteBlogRequest{
		BlogId: "5dd702170004b63fd57da98c",
	}
	res, err := cl.DeleteBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("Error happened while deleting blog: %v\n", err)
	}
	fmt.Printf("Blog has been deleted: %v\n", res.GetResult())
}

func doListBlog(cl blogpb.BlogServiceClient) {
	fmt.Println("Starting to do list blog...")
	req := &blogpb.ListBlogRequest{}
	resStream, err := cl.ListBlog(context.Background(), req)
	if err != nil {
		log.Fatalf("Error call list blog request: %v\n", err)
	}
	for {
		res, err := resStream.Recv()
		// EOF - close stream
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed reading stream from server: %v\n", err)
		}
		fmt.Printf("Response from stream server: %v\n", res.GetBlog())
	}
	fmt.Println("All blog records received.")
}