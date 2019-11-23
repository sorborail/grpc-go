package main

import (
	"context"
	"fmt"
	"github.com/sorborail/grpc-go/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

type server struct {}

var collection *mongo.Collection

type blogItem struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
	AuthorId string       `bson:"author_id"`
	Title string          `bson:"title"`
	Content string        `bson:"content"`
}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	fmt.Println("Begin create new blog....")
	blog := req.GetBlog()
	data := blogItem{
		AuthorId: blog.GetAuthorId(),
		Title: blog.GetTitle(),
		Content: blog.GetContent(),
	}
	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal error: %v", err))
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		fmt.Println("Error creating blog record!")
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot convert to OID: %v", err))
	}
	fmt.Println("New blog record has been created.")
	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:                   oid.Hex(),
			AuthorId:             blog.GetAuthorId(),
			Title:                blog.GetTitle(),
			Content:              blog.GetContent(),
		},
	}, nil
}

func (*server) ReadBlog(ctx context.Context,req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	fmt.Println("Begin request on reading blog....")
	blogId := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID"))
	}
	// create an empty struct
	data := &blogItem{}
	filter := bson.M{"_id": oid}
	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Cannot find blog with specified ID: %v", err))
	}
	fmt.Printf("Blog with id - %v has been readed.\n", blogId)
	return &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			Id:                   data.ID.Hex(),
			AuthorId:             data.AuthorId,
			Title:                data.Title,
			Content:              data.Content,
		},
	}, nil
}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	fmt.Println("Begin request on updating blog....")
	blog := req.GetBlog()

	oid, err := primitive.ObjectIDFromHex(blog.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID"))
	}
	//filter := bson.M{"_id": oid}
	//upd := bson.M{"$set", "author_id": blog.GetAuthorId(), "title": blog.GetTitle(), "content": blog.GetContent()}
	filter := bson.D{{"_id", oid}}
	update := bson.D{{"$set", bson.D{
		{"author_id", blog.GetAuthorId()},
		{"title", blog.GetTitle()},
		{"content", blog.GetContent()},
	}}}
	res, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Cannot find blog with specified ID: %v", err))
	}
	if res.ModifiedCount == 1 {
		fmt.Printf("Blog with id - %v has been updated.\n", blog.Id)
		return &blogpb.UpdateBlogResponse{Blog: blog}, nil
	} else {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot update blog with specified ID: %v", err))
	}
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	fmt.Println("Begin request on deleting blog....")
	blogId := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID"))
	}
	// create an empty struct
	filter := bson.D{{"_id", oid}}
	res, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Cannot find blog with specified ID: %v", err))
	}
	if res.DeletedCount == 1 {
		fmt.Printf("Blog with id - %v has been deleted.\n", blogId)
		return &blogpb.DeleteBlogResponse{Result: true}, nil
	} else {
		fmt.Printf("Error while Blog with id - %v delete.\n", blogId)
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot delete blog with specified ID: %v", err))
	}
}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	fmt.Println("Begin request on list blog....")
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Cannot fetch all blog records: %v", err))
	}
	defer cursor.Close(context.Background())
	var results []bson.M
	if err := cursor.All(context.Background(), &results); err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Cannot fetch all blog records: %v", err))
	}
	for _, result := range results {
		fmt.Println(result)
		err := stream.Send(&blogpb.ListBlogResponse{Blog: &blogpb.Blog{
			Id:                   result["_id"].(primitive.ObjectID).Hex(),
			AuthorId:             result["author_id"].(string),
			Title:                result["title"].(string),
			Content:              result["content"].(string),
		}})
		if err != nil {
			return status.Errorf(codes.Internal, fmt.Sprintf("Cannot send blog data to client: %v", err))
		}
		time.Sleep(1000 * time.Millisecond)
	}
	fmt.Println("Reading blog list is done!")
	return nil
}

func main() {
	// if we crash the go code, we get a file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Blog Service begin starting...")
	var opts []grpc.ServerOption
	if len(os.Args) > 1 && os.Args[1] == "TLS" {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
		if sslErr != nil {
			log.Fatalf("Error loading server certificate: %v", sslErr)
		}
		opts = append(opts, grpc.Creds(creds))
		fmt.Println("Server will be started in TLS mode...")
	}

	// Connect to MongoDB
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil { log.Fatal(err) }
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil { log.Fatal(err) }
	fmt.Println("Connected to MongoDB.")

	collection = client.Database("blog_db").Collection("blog")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	srv := grpc.NewServer(opts...)
	blogpb.RegisterBlogServiceServer(srv, &server{})

	go func() {
		fmt.Println("Blog service is started.")
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("Failed to serve server: %v", err)
		}
	}()
	// Wait for Control C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	//Block until a signal is received
	<- ch
	fmt.Println("Stopping the Blog Service...")
	srv.Stop()
	fmt.Println("Closing the Listener...")
	_ = lis.Close()
	fmt.Println("Closing MongoDB Connection...")
	_ = client.Disconnect(ctx)
	fmt.Println("End of program.")
}
