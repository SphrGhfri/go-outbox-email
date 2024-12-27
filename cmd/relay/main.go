package main

import (
	"log"
	"os"
	"os/signal"
	"outbox/config"
	"outbox/database"
	"outbox/queue"
	"outbox/shared"
	"syscall"

	"github.com/robfig/cron/v3"
)

func main() {
	config, err := config.LoadConfig("../../.env")
	if err != nil {
		log.Println("loading conf file: ", err)
	}

	// 1. Connect to Postgres
	db, err := database.NewConnection(*config)
	if err != nil {
		log.Fatal("error connecting to db")
	}

	// 2. Connect to NATS
	conn, err := queue.CreateConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Drain() // Gracefully close on exit

	// 3. Create JetStream context
	js, err := queue.CreateJetStreamContext(conn)
	if err != nil {
		log.Fatal(err)
	}

	// 4. Initialize your Outbox Processor with JetStream
	jobProcessor := shared.OutboxProcesser{
		DB:        db,
		JSContext: js,
		Subject:   "outbox.created", // or any subject pattern you want
	}

	// 5. Set up cron to run every 10s
	c := cron.New()
	_, err = c.AddFunc("@every 10s", jobProcessor.HandleOutboxMessage)
	if err != nil {
		log.Fatal("register handler error", err)
	}
	log.Println("Start processing outbox messages")
	c.Start()
	defer c.Stop()

	// 6. Wait for terminate signal
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
	<-kill
}
