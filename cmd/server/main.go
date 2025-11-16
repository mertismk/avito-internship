package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/mertismk/avito-internship/internal/database"
	"github.com/mertismk/avito-internship/internal/handler"
	"github.com/mertismk/avito-internship/internal/repository"
	"github.com/mertismk/avito-internship/internal/service"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)

	cfg := database.NewConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	repo := repository.New(db)
	svc := service.New(repo)
	h := handler.New(svc)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	h.SetupRoutes(r)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	logger.Printf("starting server on port %s", port)
	if err := r.Run(":" + port); err != nil {
		logger.Fatal(err)
	}
}
