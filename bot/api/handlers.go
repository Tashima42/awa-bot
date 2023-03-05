package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tashima42/awa-bot/bot/pkg/auth"
	"github.com/tashima42/awa-bot/bot/pkg/db"
	"net/http"
)

type Handler struct {
	repo       *db.Repo
	hashHelper *auth.HashHelper
}

func NewHandler(repo *db.Repo, hashHelper *auth.HashHelper) *Handler {
	return &Handler{
		repo:       repo,
		hashHelper: hashHelper,
	}
}

func (h *Handler) RegisterWater(c *gin.Context) {
	var registerWaterInput RegisterWaterInput
	if err := c.BindJSON(&registerWaterInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctxUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing user in context"})
		return
	}
	user := ctxUser.(db.User)
	err := h.repo.RegisterWater(c, db.Water{
		UserId: user.Id,
		Amount: registerWaterInput.Amount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, RegisterWaterOutput{Success: true})
}
