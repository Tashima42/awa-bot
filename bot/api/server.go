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
	r.Use(handler.CORSMiddleware)

	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })

	r.Use(handler.AuthMiddleware)
	r.POST("/water", handler.RegisterWater)
	r.GET("/water", handler.GetWater)
	r.GET("/whoami", handler.WhoAmI)

	r.Run(":8096")
}
