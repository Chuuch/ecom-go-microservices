package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/chuuch/product-microservice/internal/models"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	conn, err := kafka.DialContext(context.Background(), "tcp", "localhost:9092")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	w := &kafka.Writer{
		Addr:         kafka.TCP("localhost:9091", "localhost:9092", "localhost:9093"),
		Topic:        "create_product",
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: -1,
		MaxAttempts:  3,
	}

	defer w.Close()

	product := &models.Product{
		ProductID:   primitive.NewObjectID(),
		CategoryID:  primitive.NewObjectID(),
		Name:        "Test Product",
		Description: "Test Description",
		Price:       100.00,
		Quantity:    100,
		Rating:      5,
		Photos:      []string{"https://example.com/photo1.jpg", "https://example.com/photo2.jpg"},
		ImageURL:    nil,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	productBytes, err := json.Marshal(&product)
	if err != nil {
		log.Fatal(err)
	}

	msg := kafka.Message{
		Value: productBytes,
	}

	err = w.WriteMessages(context.Background(), msg)
	if err != nil {
		fmt.Println(err)
	}
}
