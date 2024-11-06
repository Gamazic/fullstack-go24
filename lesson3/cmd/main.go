package main

import (
	"lesson3/internal"
	"log"
)

func main() {
	app := internal.NewApi()
	err := app.FiberApp.Listen(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
