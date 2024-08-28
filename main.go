package main

import (
	"context"
	"flag"
	"log"

	tgClinet "go-tg/clients/telegram"
	event_consumer "go-tg/consumer/event-consumer"
	"go-tg/events/telegram"
	// "go-tg/storage/files"
	"go-tg/storage/sqlite"
)

const (
	tgBotHost = "api.telegram.org"
	// storagePath = "files_storage"
	sqliteStoragePath = "data/sqlite/storage.db"
	batchSize = 100
)

func main () {
	// s := files.New(storagePath)
	s, err := sqlite.New(sqliteStoragePath)
	if err != nil {
		log.Fatal("Can't conntect to storage: %w", err)
	}

	err = s.Init(context.TODO())

	if err != nil {
		log.Fatal("Can't init storage: %w", err)
	}

	eventsProcessor := telegram.New(
		tgClinet.New(tgBotHost, mustToken()),
		s,
	)

	log.Print("service started")

	consumer := event_consumer.New(eventsProcessor, eventsProcessor, batchSize)
	if err:= consumer.Start(); err != nil {
		log.Fatal("Service is stopped")
	}
}

func mustToken() string {
	token := flag.String(
		"token-bot",
		"",
		"telegram bot token",
	)
	flag.Parse()
	if *token == "" {
		log.Fatal("Token not found")
	}
	return *token
}