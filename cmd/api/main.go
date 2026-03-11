package main

import (
	"log"

	"github.com/Laiirton/time-control-api-go/internal/config"
	"github.com/Laiirton/time-control-api-go/internal/database"
	"github.com/Laiirton/time-control-api-go/internal/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Erro ao carregar configuração: ", err)
	}

	db, err := database.Connect(cfg.DBURL)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco de dados: ", err)
	}
	defer db.Close()

	log.Println("Conectado ao banco de dados com sucesso")

	r := gin.Default()
	routes.Setup(r, db, cfg.JWTSecret)

	log.Printf("Servidor rodando na porta %s", cfg.APIPort)
	if err := r.Run(":" + cfg.APIPort); err != nil {
		log.Fatal("Erro ao iniciar servidor: ", err)
	}
}
