package main

import (
	"fmt"
	"log"
	"net"
	"outbox/config"
	"outbox/database"
	"outbox/notification"
	"outbox/pb"
	"outbox/shared"

	"google.golang.org/grpc"
)

func main() {
	config, err := config.LoadConfig("../../.env")
	if err != nil {
		log.Println("loading config file: ", err)
	}

	db, err := database.NewConnection(*config)
	if err != nil {
		log.Fatal("error connecting to db")
	}

	if err := db.AutoMigrate(&shared.OutBoxMessage{}); err != nil {
		log.Fatal("migrate error - ", err)
	}

	grpcServer := grpc.NewServer()

	svc := &notification.Service{DB: db}
	pb.RegisterNotificationServiceServer(grpcServer, svc)

	address := fmt.Sprintf(":%d", config.Port)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", config.Port, err)
	}
	log.Printf("Starting gRPC server on :%d...\n", config.Port)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
