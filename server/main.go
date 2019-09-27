package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	productpb "grpc-mongo-crud/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ProductServiceServer struct
type ProductServiceServer struct{}

// CreateProduct method
func (s *ProductServiceServer) CreateProduct(ctx context.Context, req *productpb.CreateProductReq) (*productpb.CreateProductRes, error) {
	// Access reqProduct struct and convert to BSON
	product := req.GetProduct()
	data := ProductItem{
		Category: product.GetCategory(),
		Title:    product.GetTitle(),
		Price:    product.GetPrice(),
		Quantity: product.GetQuantity(),
	}

	// Add data to database
	result, err := productdb.InsertOne(mongoCtx, data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internet error: %v", err),
		)
	}

	// Add id to the product
	oid := result.InsertedID.(primitive.ObjectID)
	product.Id = oid.Hex()

	// Returning product in a CreateProductRes type
	return &productpb.CreateProductRes{Product: product}, nil
}

// GetProduct method
func (s *ProductServiceServer) GetProduct(ctx context.Context, req *productpb.GetProductReq) (*productpb.GetProductRes, error) {
	// convert string ID to mongoDb ObjectId
	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	}

	result := productdb.FindOne(ctx, bson.M{"_id": oid})

	// Empty product to write decode result
	data := ProductItem{}

	// decode and write to data
	if err := result.Decode(&data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not find product with Object Id %s: %v", req.GetId(), err))
	}

	response := &productpb.GetProductRes{
		Product: &productpb.Product{
			Id:       oid.Hex(),
			Category: data.Category,
			Title:    data.Title,
			Price:    data.Price,
			Quantity: data.Quantity,
		},
	}

	return response, nil
}

// UpdateProduct method
func (s *ProductServiceServer) UpdateProduct(ctx context.Context, req *productpb.UpdateProductReq) (*productpb.UpdateProductRes, error) {
	// Product data from request
	product := req.GetProduct()

	// Convert id string to MongoDb ObjectId
	oid, err := primitive.ObjectIDFromHex(product.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Could not convert supplied id to a MongoDb ObjectId %v", err),
		)
	}

	// Convert data to be updated into a unordered Bson document
	update := bson.M{
		"category": product.GetCategory(),
		"title":    product.GetTitle(),
		"price":    product.GetPrice(),
		"quantity": product.GetQuantity(),
	}

	// Convert oid into unordered to search bson document by id
	filter := bson.M{"_id": oid}

	// Return encoded bson result
	result := productdb.FindOneAndUpdate(ctx, filter, bson.M{"$set": update}, options.FindOneAndUpdate().SetReturnDocument(1))

	// Decode result and write it to encode
	decode := ProductItem{}
	err = result.Decode(&decode)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not find product with supplied ID: %v", err))
	}

	return &productpb.UpdateProductRes{
		Product: &productpb.Product{
			Id:       decode.ID.Hex(),
			Category: decode.Category,
			Title:    decode.Title,
			Price:    decode.Price,
			Quantity: decode.Quantity,
		},
	}, nil
}

// DeleteProduct method
func (s *ProductServiceServer) DeleteProduct(ctx context.Context, req *productpb.DeleteProductReq) (*productpb.DeleteProductRes, error) {
	// convert string ID to mongoDb ObjectId
	oid, err := primitive.ObjectIDFromHex(req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Could not convert to ObjectId: %v", err))
	}

	// Delete product
	_, err = productdb.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Could not delete product with Object Id %s: %v", req.GetId(), err))
	}

	return &productpb.DeleteProductRes{
		Success: true,
	}, nil
}

// GetAllProducts method
func (s *ProductServiceServer) GetAllProducts(req *productpb.GetAllProductsReq, stream productpb.ProductService_GetAllProducts) error {
	// Initiante ProductItem type to write decoded data to
	data := &ProductItem{}

	// return a cursor for empty query
	cursor, err := productdb.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Unknow internal error %v", err))
	}

	// Close cursor when there is no more data to stream
	defer cursor.Close(context.Background())

	// return boolean, if false cursor.Close()
	for cursor.Next(context.Background()) {
		err := cursor.Decode(data)
		if err != nil {
			return nil, status.Errorf(codes.Unavailable, fmt.Sprintf("Could not decode data: %v", err))
		}

		// If no error send product over stream
		stream.Send(&productpb.GetAllProductsRes{
			Product: &productpb.Product{
				Id:       data.ID.Hex(),
				Category: data.Category,
				Title:    data.Title,
				Price:    data.Price,
				Quantity: data.Quantity,
			},
		})
	}

	// Check if cursor has any errors
	if err := cursor.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Unknow cursor error: %v", err))
	}

	return nil, nil
}

// ProductItem struct
type ProductItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Category string             `bson:"category"`
	Title    string             `bson:"title"`
	Price    string             `bson:"price"`
	Quantity string             `bson:"quantity"`
}

var db *mongo.Client
var productdb *mongo.Collection
var mongoCtx context.Context

const port = ":1337"

func main() {
	// Starting server
	log.SetFlags(log.Llongfile | log.Lshortfile)
	fmt.Println("Starting server on port :1337...")

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Unable to listen po port "+port+": %v", err)
	}

	// Starting gRPC server, options
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	srv := &ProductServiceServer{}
	productpb.RegisterProductServiceServer(s, srv)

	// Connecting to MongoDb client and check if connection successful
	fmt.Println("Connecting to MongoDb...")
	mongoCtx = context.Background()

	db, err = mongo.Connect(mongoCtx, options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping(mongoCtx, nil)
	if err != nil {
		log.Fatalf("Could not connect to MongoDB: %v\n", err)
	} else {
		fmt.Println("Connected to Mongodb")
	}

	productdb = db.Database("mydb").Collection("products")

	// Start server in a child routine, SHUTDOWN hook
	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	fmt.Println("Server succesfully started on port ", port)

	c := make(chan os.Signal)

	signal.Notify(c, os.Interrupt)

	// As long as user dosent press CTRL+C main routine keeeps running
	<-c

	fmt.Println("\nStopping server...")
	s.Stop()
	listener.Close()
	fmt.Println("Closing MongoDb connection")
	db.Disconnect(mongoCtx)
	fmt.Println("Done.")
}
