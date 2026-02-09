package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/hamba/avro"
)

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

type OrderCreated struct {
	OrderID          string `avro:"order_id"`
	StoreID          string `avro:"store_id"`
	TotalAmountCents int    `avro:"total_amount_cents"`
	Status           string `avro:"status"`
	CreatedAt        int64  `avro:"created_at"`
}

func main() {
	// 1. Parse Schema
	schema, err := avro.Parse(schemaStr)
	if err != nil {
		log.Fatal("Failed to parse schema:", err)
	}

	// 2. Init Consumer
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "kafka:29092",
		"group.id":          "kitchen-service-group",
		"auto.offset.reset": "earliest",
		"enable.auto.commit": false, // Manual commit for safety
	})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	topic := "global.orders.v1"
	c.SubscribeTopics([]string{topic}, nil)

	log.Println("Kitchen Service Started. Waiting for orders...")

	// 3. Consume Loop
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	run := true
	for run {
		select {
		case sig := <-sigchan:
			log.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				// Deserialize
				var order OrderCreated
				err = avro.Unmarshal(schema, e.Value, &order)
				if err != nil {
					log.Printf("Error deserializing: %v\n", err)
					continue
				}

				// "Process" the order
				fmt.Printf("👨‍🍳 KITCHEN: Preparing Order %s (Store: %s, Total: $%.2f)\n", 
					order.OrderID, order.StoreID, float64(order.TotalAmountCents)/100.0)

				// Commit offset
				_, err = c.CommitMessage(e)
				if err != nil {
					log.Printf("Error committing offset: %v\n", err)
				}
				
			case kafka.Error:
				log.Printf("%% Error: %v\n", e)
			}
		}
	}
}
