package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/joho/godotenv"
	"github.com/piatoss3612/trade-watcher/internal/binance"
	"github.com/piatoss3612/trade-watcher/internal/producer"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// kafkaHost := os.Getenv("KAFKA_HOST")
	kafkaHost := "localhost"
	kafkaPort := os.Getenv("KAFKA_PORT")
	topics := strings.Split(os.Getenv("TICKERS"), ",")

	sub, err := binance.NewSubscriber(context.Background())
	if err != nil {
		log.Fatalf("Error creating subscriber: %v", err)
	}
	defer sub.Close()

	err = sub.Subscribe(topics)
	if err != nil {
		log.Fatalf("Error subscribing: %v", err)
	}

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("%s:%s", kafkaHost, kafkaPort),
	})
	if err != nil {
		log.Fatalf("Error creating producer: %v", err)
	}

	prod := producer.New(p, os.Getenv("KAFKA_TOPIC"), true)

	stop := make(chan struct{})
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer close(stop)

		go prod.ListenAndProduce(sub)

		<-sigChan
	}()

	<-stop

	log.Println("Shutting down...")

	prod.Close()

	err = sub.UnSubscribe()
	if err != nil {
		log.Fatalf("Error unsubscribing: %v", err)
	}

	err = sub.Close()
	if err != nil {
		log.Fatalf("Error closing subscriber: %v", err)
	}

	log.Println("Done")
}
