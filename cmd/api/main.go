package main

import (
	"log"
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"Sportify/internal/server"
)

func main() {
	mongoClient, err := mongo.Connect(options.Client().ApplyURI("mongodb://admin:admin@mongo:27017"))
	if err != nil {
		log.Fatal("Greška prilikom konektovanja na MongoDB bazu podataka: ", err)
	}

	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Greška prilikom zatvaranja MongoDB konekcije: %v\n", err)
		}
	}()

	server, port := server.NoviServer(mongoClient)
	log.Printf("Server osluškuje na adresi: %v\n", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("ListenAndServe greška: ", err)
	}
}
