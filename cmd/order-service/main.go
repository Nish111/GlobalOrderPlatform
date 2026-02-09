package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"
	"github.com/hamba/avro"
	"github.com/google/uuid"
)

// Schema definition (embedded for simplicity, in prod load from file/registry)
const schemaStr = `{
  "namespace": "com.mcdonalds.orders",
  "type": "record",
  "name": "OrderCreated",
  "fields": [
    {"name": "order_id", "type": "string"},
    {"name": "store_id", "type": "string"},
    {"name": "total_amount_cents", "type": "int"},
    {"name": "status", "type": "string"}, 
    {"name": "created_at", "type": "long"}
  ]
}`

type OrderRequest struct {
	StoreID          string `json:"store_id" binding:"required"`
	TotalAmountCents int    `json:"total_amount_cents" binding:"required"`
}

type OrderCreated struct {
	OrderID          string `avro:"order_id"`
	StoreID          string `avro:"store_id"`
	TotalAmountCents int    `avro:"total_amount_cents"`
	Status           string `avro:"status"`
	CreatedAt        int64  `avro:"created_at"`
}

var (
	producer *kafka.Producer
	schema   avro.Schema
	topic    = "global.orders.v1"
)

func main() {
	var err error
	// 1. Parse Schema
	schema, err = avro.Parse(schemaStr)
	if err != nil {
		log.Fatal("Failed to parse schema:", err)
	}

	// 2. Init Kafka Producer
	producer, err = kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:29092",
		"acks":              "all",
		"enable.idempotence": true,
	})
	if err != nil {
		log.Fatal("Failed to create producer:", err)
	}
	defer producer.Close()

	// 3. Start HTTP Server
	r := gin.Default()
	r.POST("/orders", createOrder)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	log.Println("Order Service listening on :8080")
	r.Run(":8080")
}

func createOrder(c *gin.Context) {
	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orderID := uuid.New().String()
	event := OrderCreated{
		OrderID:          orderID,
		StoreID:          req.StoreID,
		TotalAmountCents: req.TotalAmountCents,
		Status:           "NEW",
		CreatedAt:        time.Now().UnixMilli(),
	}

	// Avro Serialize
	data, err := avro.Marshal(schema, event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "serialization failed"})
		return
	}

	// Produce to Kafka
	deliveryChan := make(chan kafka.Event)
	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(orderID), // Partition by OrderID
		Value:          data,
	}, deliveryChan)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "kafka produce failed"})
		return
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": m.TopicPartition.Error.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"order_id": orderID,
		"status":   "accepted",
		"offset":   m.TopicPartition.Offset,
	})
}
