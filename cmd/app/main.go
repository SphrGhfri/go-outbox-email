package main

import (
	"log"
	"outbox/config"
	"outbox/database"
	"outbox/notification"
	"outbox/shared"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
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

	if err := db.AutoMigrate(&notification.Notification{}, &shared.OutBoxMessage{}); err != nil {
		log.Fatal("migrate error - ", err)
	}

	notificationHandler := notification.Handler{DB: db}

	app := fiber.New()

	app.Use(logger.New())

	app.Post("/notifications", notificationHandler.Add)

	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
