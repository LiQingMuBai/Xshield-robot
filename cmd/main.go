package main

import (
	"log"
	"ushield_bot/internal/bootstrap"
)

func main() {
	application, err := bootstrap.BuildApp()
	if err != nil {
		log.Fatalf("build app err: %v", err)
	}

	if err := application.Run(processUpdate); err != nil {
		log.Fatalf("run app err: %v", err)
	}
}
