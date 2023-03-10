package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tashima42/awa-bot/bot/pkg/auth"
	"github.com/tashima42/awa-bot/bot/pkg/db"
	"net/http"
	"strconv"
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

// POST /register
// register user water log
func (h *Handler) RegisterWater(c *gin.Context) {
	var registerWaterInput RegisterWaterInput
	if err := c.ShouldBindJSON(&registerWaterInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctxUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing user in context"})
		return
	}
	user := ctxUser.(*db.User)
	err := h.repo.RegisterWater(c, db.Water{
		UserId: user.Id,
		Amount: *registerWaterInput.Amount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, RegisterWaterOutput{Success: true})
}

// GET /water
// get user water list ordered by latest
func (h *Handler) GetWater(c *gin.Context) {
	var limit, skip int
	ctxUser, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing user in context"})
		return
	}
	user := ctxUser.(*db.User)
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "limit must be an integer"})
		return
	}
	skip, err = strconv.Atoi(c.Query("skip"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "skip must be an integer"})
		return
	}
	watersPointers, total, err := h.repo.GetUserWaterPaginated(c, user.Id, limit, skip)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	waters := make([]db.Water, len(watersPointers))
	for i, water := range watersPointers {
		waters[i] = *water
	}
	c.JSON(http.StatusOK, GetWaterOutput{Waters: waters, Total: total})
}
