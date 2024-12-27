package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"

	"outbox/config"
	"outbox/email"
	"outbox/queue"
)

type OutboxEvent struct {
	// ID        string `json:"id"`
	EventName string `json:"event_name"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Message   string `json:"message"`
}

func main() {
	// 1) Load config (so we have SMTP settings, etc.)
	cfg, err := config.LoadConfig("../../.env")
	if err != nil {
		log.Println("loading conf file: ", err)
	}

	// 2) Create an email service instance
	emailService := email.NewEmailService(cfg)

	// 3) Connect to NATS
	nc, err := queue.CreateConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()

	// 4) Create JetStream context
	js, err := queue.CreateJetStreamContext(nc)
	if err != nil {
		log.Fatal(err)
	}

	// 5) Subscribe to "outbox.*"
	sub, err := js.Subscribe("outbox.*", func(msg *nats.Msg) {
		var evt OutboxEvent
		if err := json.Unmarshal(msg.Data, &evt); err != nil {
			log.Println("Handle message error: ", string(msg.Data))
			log.Println("ERR:", err)
			_ = msg.Nak() // Negative-ack
			return
		}
		log.Printf("Sending Email [%s] - Emal: '%s'", evt.EventName, evt.Email)

		// (A) Prepare data for the template
		emailData := map[string]interface{}{
			"Message": evt.Message,
			"Name":    evt.Name,
		}

		// (B) Send the email via emailService
		if err := emailService.SendEmail(
			[]string{evt.Email},
			"Hello From Gholi",
			"notif.html",
			emailData,
		); err != nil {
			log.Println("Failed to send email:", err)
			_ = msg.Nak()
			return
		}

		log.Printf("Email sent successfully to %s for event %s\n", evt.Email, evt.EventName)

		// If success, Ack the message
		_ = msg.Ack()

	}, nats.Durable("WORKER"), nats.ManualAck())
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	// 6) Wait for signals
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
	<-kill
}
