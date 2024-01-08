package producer

import (
	"crypto/sha256"
	"log"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Listener interface {
	Listen() (<-chan []byte, <-chan error)
}

type Producer struct {
	producer *kafka.Producer
	topic    string
	logging  bool

	deliveryChan chan kafka.Event
	closeChan    chan struct{}
}

func New(producer *kafka.Producer, topic string, logging bool) *Producer {
	return &Producer{
		producer:     producer,
		topic:        topic,
		logging:      logging,
		deliveryChan: make(chan kafka.Event),
		closeChan:    make(chan struct{}),
	}
}

func (p *Producer) Close() {
	close(p.closeChan)
	p.producer.Close()
}

func (p *Producer) ListenAndProduce(l Listener) {
	payloads, errs := l.Listen()

	go p.watchEvent()

	for {
		select {
		case <-p.closeChan:
			return
		case payload := <-payloads:
			p.produce(payload)
		case err := <-errs:
			if p.logging {
				log.Printf("Error reading message: %v", err)
			}
		}
	}
}

func (p *Producer) produce(payload []byte) {
	h := sha256.New()
	h.Write(payload)
	checksum := h.Sum(nil)[:4]

	err := p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &p.topic,
			Partition: kafka.PartitionAny,
		},
		Key:   checksum,
		Value: payload,
	}, p.deliveryChan)
	if err != nil {
		if p.logging {
			log.Printf("Error producing message: %v", err)
		}
	}

}

func (p *Producer) watchEvent() {
	for e := range p.deliveryChan {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				if p.logging {
					log.Printf("Delivery failed: %v", ev.TopicPartition.Error)
				}
			} else {
				if p.logging {
					log.Printf("Delivered message to topic %s [%d] at offset %v\n", *ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
				}
			}
		}
	}
}
