package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tashima42/awa-bot/bot/pkg/auth"
	"github.com/tashima42/awa-bot/bot/pkg/db"
	"net/http"
)

func Serve(repo *db.Repo, hashHelper *auth.HashHelper) {
	handler := NewHandler(repo, hashHelper)

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(handler.AuthMiddleware)

	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })
	r.POST("/water", handler.RegisterWater)

	r.Run(":8096")
}
