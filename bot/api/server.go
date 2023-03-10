package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tashima42/awa-bot/bot/pkg/auth"
	"github.com/tashima42/awa-bot/bot/pkg/db"
	"net/http"
)

func Serve(repo *db.Repo, hashHelper *auth.HashHelper, jwtHelper *auth.JWTHelper) {
	handler := NewHandler(repo, hashHelper, jwtHelper)

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(handler.CORSMiddleware)

	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })
	r.POST("/login", handler.Login)

	r.Use(handler.AuthMiddleware)
	r.POST("/water", handler.RegisterWater)
	r.GET("/water", handler.GetWater)

	r.Run(":8096")
}
