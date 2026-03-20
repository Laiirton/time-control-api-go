package main

import (
	"log"

	"github.com/Laiirton/time-control-api-go/internal/config"
	"github.com/Laiirton/time-control-api-go/internal/database"
	"github.com/Laiirton/time-control-api-go/internal/routes"
	"github.com/gin-contrib/cors"
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

	if err := database.RunMigrations(db); err != nil {
		log.Fatal("Erro ao executar migrações do banco de dados: ", err)
	}
	log.Println("Migrações do banco de dados executadas com sucesso")

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:8081"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	routes.Setup(r, db, cfg)

	log.Printf("Servidor rodando na porta %s", cfg.APIPort)
	if err := r.Run("0.0.0.0:" + cfg.APIPort); err != nil {
		log.Fatal("Erro ao iniciar servidor: ", err)
	}
}
