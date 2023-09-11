package handler

import (
	"github.com/gin-gonic/gin"
	"proxy/Proxy/internal/pkg/api"
)

type Handler struct {
	usecase api.Usecase
}

func NewHandler(usecase api.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (handler *Handler) SetupRoutes() *gin.Engine {
	router := gin.New()
	api := router.Group("/")
	{
		api.GET("/requests", handler.AllRequests)
		api.GET("/requests/{id}", handler.GetRequest)
		api.GET("/repeat/{id}", handler.RepeatRequest)
		api.GET("/scan/{id}", handler.Scan)
	}

	return router
}

func newErrorResponse(c *gin.Context, statusCode int, details string) {
	c.AbortWithStatusJSON(statusCode, details)
}

func (handler *Handler) AllRequests(c *gin.Context) {

}

func (handler *Handler) GetRequest(c *gin.Context) {

}

func (handler *Handler) RepeatRequest(c *gin.Context) {

}

func (handler *Handler) Scan(c *gin.Context) {

}
