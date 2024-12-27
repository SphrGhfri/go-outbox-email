package queue

import (
	"fmt"
	"os"

	"github.com/nats-io/nats.go"
)

func CreateConnection() (*nats.Conn, error) {
	// We assume we have environment variables for NATS similar to RABBITMQ
	// e.g., NATS_HOST, NATS_PORT, etc. Adjust to your real environment:
	url := fmt.Sprintf("nats://%s:%s", os.Getenv("NATS_HOST"), os.Getenv("NATS_PORT"))
	return nats.Connect(url)
}

func CreateJetStreamContext(conn *nats.Conn) (nats.JetStreamContext, error) {
	// Create a JetStream context:
	js, err := conn.JetStream()
	if err != nil {
		return nil, err
	}

	// Create a stream, e.g. "OUTBOX"
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     "OUTBOX",
		Subjects: []string{"outbox.*"},
	})
	if err != nil {
		return nil, err
	}

	return js, nil
}
