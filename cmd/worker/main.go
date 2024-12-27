package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"outbox/queue"
	"syscall"

	"github.com/nats-io/nats.go"
	"gorm.io/datatypes"
)

type OutboxEvent struct {
	ID        string         `json:"id"`
	EventName string         `json:"event_name"`
	Payload   datatypes.JSON `json:"payload"`
}

func main() {
	// 1. Connect to NATS
	nc, err := queue.CreateConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()

	// 2. Create JetStream context
	js, err := queue.CreateJetStreamContext(nc)
	if err != nil {
		log.Fatal(err)
	}

	// 3. Subscribe to the stream/subject
	sub, err := js.Subscribe("outbox.*", func(msg *nats.Msg) {
		var evt OutboxEvent
		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			log.Println("Handle message error: ", string(msg.Data))
			log.Println("ERR:", err)
			// NACK or Ack?
			_ = msg.Nak()
			return
		}
		log.Printf("Handling [%s] - Payload: '%s'", evt.EventName, evt.Payload)

		// If success, Ack the message
		_ = msg.Ack()
	}, nats.Durable("WORKER"), nats.ManualAck())

	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	// 4. Wait for signals
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
	<-kill
}
