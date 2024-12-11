package main

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"io"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Configuration constants
const (
	mongoURI       = "mongodb://root:password@mongo:27017"
	databaseName   = "image_service"
	collectionName = "images"
	s3Endpoint     = "http://minio:9000"
	s3BucketName   = "uploads"
	s3AccessKey    = "minioadmin"
	s3SecretKey    = "minioadmin"
	appPort        = ":8000"
)

// Global variables
var (
	mongoColl *mongo.Collection
	s3Client  *s3.S3
)

// ImageRecord represents a record in MongoDB
type ImageRecord struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	S3URL     string             `bson:"s3_url"`
	CreatedAt time.Time          `bson:"created_at"`
	ImageMeta ImageMeta          `bson:"image_meta"`
}

type ImageMeta struct {
	Size int `bson:"size"`
}

func main() {
	// Initialize MongoDB
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	mongoColl = mongoClient.Database(databaseName).Collection(collectionName)

	// Initialize MinIO (S3)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"), // MinIO uses "us-east-1" as a default
		Credentials: credentials.NewStaticCredentials(
			s3AccessKey, // Access Key
			s3SecretKey, // Secret Key
			""),
		Endpoint:         aws.String(s3Endpoint),
		S3ForcePathStyle: aws.Bool(true), // Path-style addressing for MinIO
		DisableSSL:       aws.Bool(true),
	})
	if err != nil {
		log.Fatalf("Failed to initialize S3 client: %v", err)
	}
	s3Client = s3.New(sess)

	// Create the Fiber app
	app := fiber.New()
	app.Use(logger.New())

	// Routes
	app.Post("/upload", handleUpload)
	app.Get("/image/:id", handleGetImage)

	log.Printf("Server is running on http://localhost%s", appPort)
	log.Fatal(app.Listen(appPort))
}

// handleUpload handles the image upload, stores it in S3, and saves the link in MongoDB
func handleUpload(c *fiber.Ctx) error {
	file, err := c.FormFile("image")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Failed to get image: %v", err))
	}

	// Open the file
	fileContent, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to open file: %v", err))
	}
	defer fileContent.Close()

	// Generate a unique file name
	fileName := file.Filename
	s3Key := fmt.Sprintf("uploads/%s", fileName)

	// Upload the file to MinIO (S3)
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(s3Key),
		Body:   fileContent,
		ACL:    aws.String(s3.BucketCannedACLPublicRead),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to upload to S3: %v", err))
	}

	// S3 URL
	s3URL := fmt.Sprintf("%s/%s/%s", s3Endpoint, s3BucketName, s3Key)

	// Save the record in MongoDB
	imageRecord := ImageRecord{
		S3URL:     s3URL,
		CreatedAt: time.Now(),
	}
	result, err := mongoColl.InsertOne(context.TODO(), imageRecord) // bson.D{"key": "value", "ley1": 12312}.)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to save to database: %v", err))
	}

	// Return the ID of the new record
	id := result.InsertedID.(primitive.ObjectID)
	return c.JSON(fiber.Map{
		"id": id.Hex(),
	})
}

func handleGetImage(c *fiber.Ctx) error {
	id := c.Params("id")

	// Parse the ID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("Invalid ID format: %v", err))
	}

	// Fetch the record from MongoDB
	var imageRecord ImageRecord
	err = mongoColl.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&imageRecord)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fiber.NewError(fiber.StatusNotFound, fmt.Sprintf("Image not found: %v", err))
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to fetch record: %v", err))
	}

	// Parse the S3 key from the URL (assuming standard format)
	s3Key := imageRecord.S3URL[len(s3Endpoint)+len(s3BucketName)+2:]

	// Get the file from S3
	output, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to retrieve file from S3: %v", err))
	}
	defer output.Body.Close()

	// Set response headers
	c.Set("Content-Type", "image/jpeg")
	c.Set("Content-Length", fmt.Sprintf("%d", aws.Int64Value(output.ContentLength)))

	// Stream file content to response
	if _, err := io.Copy(c.Response().BodyWriter(), output.Body); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to stream file to response: %v", err))
	}

	return nil
}
