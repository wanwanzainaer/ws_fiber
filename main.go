package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
	handlers "wsfiber/routes"
)
func main(){
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, world")
	})
	app.Get("/ws",handlers.WsHandler)
	go handlers.ListenToWsChannel()
	log.Fatal(app.Listen(":3000"))
}
