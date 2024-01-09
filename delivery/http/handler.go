package http

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/learn/letter-generator/usecase"
)

type Handler struct {
	GenerateLetterUsecase usecase.GenerateLetterUsecase
}

func New(generateLetterUsecase usecase.GenerateLetterUsecase) Handler {
	return Handler{GenerateLetterUsecase: generateLetterUsecase}
}

func (h Handler) InitRouter(router *gin.Engine) {
	router.POST("/generate-letter", h.GenerateLetter)
}

func (h Handler) GenerateLetter(c *gin.Context) {
	var input usecase.GenerateLetterRequest

	c.BindJSON(&input)

	res, err := h.GenerateLetterUsecase.GenerateLetter(input)
	if err != nil {
		log.Printf("[handler] error while do GenerateLetter %+v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"Message": "Internal Server Error", "InternalMessage": err.Error()})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{"Message": "Success", "Response": res})
	c.Abort()
}
