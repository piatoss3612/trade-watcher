package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/joho/godotenv"
	"github.com/piatoss3612/trade-watcher/internal/binance"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaHost := "localhost"
	kafkaPort := os.Getenv("KAFKA_PORT")

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("%s:%s", kafkaHost, kafkaPort),
		"group.id":          "trade-watcher",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		log.Fatalf("Error creating consumer: %v", err)
	}

	topic := os.Getenv("KAFKA_TOPIC")

	err = c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Fatalf("Error subscribing to topic: %v", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	run := true

	for run {
		select {
		case sig := <-sigChan:
			log.Printf("Caught signal %v: terminating", sig)
			run = false
		default:
			msg, err := c.ReadMessage(100 * time.Millisecond)
			if err != nil && err.(kafka.Error).Code() != kafka.ErrTimedOut {
				log.Printf("Error reading message: %v", err)
			}

			if msg == nil {
				continue
			}

			var trade binance.AggregateTrade

			err = json.Unmarshal(msg.Value, &trade)
			if err != nil {
				log.Printf("Error unmarshalling message: %v", err)
			}

			log.Printf("Message on %s: %s", msg.TopicPartition, string(msg.Value))
		}
	}
}
