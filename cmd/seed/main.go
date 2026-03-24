package main

import (
	"flag"
	"log"

	"github.com/Laiirton/time-control-api-go/internal/config"
	"github.com/Laiirton/time-control-api-go/internal/database"
)

func main() {
	count := flag.Int("count", 10, "number of random users to create")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration: ", err)
	}

	db, err := database.Connect(cfg.DBURL)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer db.Close()

	inserted, err := database.SeedUsers(db, *count)
	if err != nil {
		log.Fatal("Failed to run user seeder: ", err)
	}

	log.Printf("User seeder executed successfully. Users inserted: %d", inserted)
}
