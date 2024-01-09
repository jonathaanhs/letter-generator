package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/learn/letter-generator/config"
	httpDelivery "github.com/learn/letter-generator/delivery/http"
	"github.com/learn/letter-generator/usecase"
)

func main() {

	cfg := config.Init()

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	generateLetterUsecase := usecase.New(cfg.Sheets, cfg.Docs, cfg.Drive)

	handler := httpDelivery.New(generateLetterUsecase)

	handler.InitRouter(router)

	log.Println("Starting HTTP server")

	err := router.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
