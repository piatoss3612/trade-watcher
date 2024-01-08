package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/url"

	"github.com/gorilla/websocket"
)

type Subscriber struct {
	conn      *websocket.Conn
	id        int64
	topics    []string
	unsubChan chan struct{}
}

func NewSubscriber(ctx context.Context) (*Subscriber, error) {
	u := url.URL{Scheme: "wss", Host: "stream.binance.com:443", Path: "/ws"}

	c, resp, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 101 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return &Subscriber{
		conn:      c,
		id:        rand.Int63(),
		unsubChan: make(chan struct{}),
	}, nil
}

func (s *Subscriber) Close() error {
	return s.conn.Close()
}

func (s *Subscriber) Subscribe(topics []string) error {
	conn := s.conn

	conn.SetPongHandler(s.pongHandler)

	tradeTopics := make([]string, 0, len(topics))
	for _, topic := range topics {
		tradeTopics = append(tradeTopics, topic+"@"+"aggTrade")
	}

	s.topics = tradeTopics

	message := Request{
		Id:     s.id,
		Method: "SUBSCRIBE",
		Params: tradeTopics,
	}

	err := s.sendMsg(&message)
	if err != nil {
		return err
	}

	var resp Response

	err = s.readMsg(&resp)
	if err != nil {
		return err
	}

	if resp.Id != s.id {
		return fmt.Errorf("unexpected id: %d", resp.Id)
	}

	return nil
}

func (s *Subscriber) UnSubscribe() error {
	close(s.unsubChan)

	message := Request{
		Id:     s.id,
		Method: "UNSUBSCRIBE",
		Params: s.topics,
	}

	return s.sendMsg(&message)
}

func (s *Subscriber) Listen() (<-chan []byte, <-chan error) {
	errorChan := make(chan error)
	payloadChan := make(chan []byte)

	go func() {
		defer func() {
			close(errorChan)
			close(payloadChan)
		}()

		for {
			select {
			case <-s.unsubChan:
				return
			default:
				_, payload, err := s.conn.ReadMessage()
				if err != nil {
					errorChan <- err
				}

				payloadChan <- payload
			}
		}
	}()

	return payloadChan, errorChan
}

func (s *Subscriber) sendMsg(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return s.conn.WriteMessage(websocket.TextMessage, b)
}

func (s *Subscriber) readMsg(t interface{}) error {
	_, payload, err := s.conn.ReadMessage()
	if err != nil {
		return err
	}

	err = json.Unmarshal(payload, t)
	if err != nil {
		return err
	}

	return nil
}

func (s *Subscriber) pongHandler(appData string) error {
	pingFrame := []byte{1, 2, 3, 4, 5}
	return s.conn.WriteMessage(websocket.PongMessage, pingFrame)
}
