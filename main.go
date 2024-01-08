package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/piatoss3612/trade-watcher/binance"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

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

	tickets, errs := sub.Listen()

	stop := make(chan struct{})
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-sigChan:
				close(stop)
				return
			case ticket := <-tickets:
				log.Printf("Ticket: %+v", ticket)
			case err := <-errs:
				log.Printf("Error: %v", err)
			}
		}
	}()

	<-stop

	log.Println("Shutting down...")

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
