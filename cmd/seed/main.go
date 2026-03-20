package main

import (
	"flag"
	"log"

	"github.com/Laiirton/time-control-api-go/internal/config"
	"github.com/Laiirton/time-control-api-go/internal/database"
)

func main() {
	count := flag.Int("count", 10, "quantidade de usuários aleatórios para criar")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Erro ao carregar configuração: ", err)
	}

	db, err := database.Connect(cfg.DBURL)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco de dados: ", err)
	}
	defer db.Close()

	inserted, err := database.SeedUsers(db, *count)
	if err != nil {
		log.Fatal("Erro ao executar seeder de usuários: ", err)
	}

	log.Printf("Seeder de usuários executado com sucesso. Usuários inseridos: %d", inserted)
}
