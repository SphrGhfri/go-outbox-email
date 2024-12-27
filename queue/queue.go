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
		Name:      "OUTBOX",
		Subjects:  []string{"outbox.*"},
		Retention: nats.WorkQueuePolicy,
	})
	if err != nil {
		return nil, err
	}

	_, err = js.ConsumerInfo("OUTBOX", "WORKER")
	if err != nil {
		// Means consumer doesn't exist, so let's create it (push-based).
		_, err = js.AddConsumer("OUTBOX", &nats.ConsumerConfig{
			Durable:        "WORKER",
			AckPolicy:      nats.AckExplicitPolicy,
			MaxDeliver:     5,
			DeliverSubject: "push.outbox", // must include for push mode
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Consumer already exists, so update it (keep it push-based).
		_, err = js.UpdateConsumer("OUTBOX", &nats.ConsumerConfig{
			Durable:        "WORKER",
			AckPolicy:      nats.AckExplicitPolicy,
			MaxDeliver:     5,
			DeliverSubject: "push.outbox", // keep the same deliver subject
		})
		if err != nil {
			return nil, err
		}
	}

	return js, nil
}
