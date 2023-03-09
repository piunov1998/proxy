package main

import (
	"log"
	"ws_rest_wrapper/internal/app"
)

func main() {
	application := app.App{}
	log.Fatalln(application.Run())
}
