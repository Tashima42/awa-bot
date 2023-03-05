package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *Handler) AuthMiddleware(c *gin.Context) {
	var apiKey, userID string
	var err error

	apiKey = c.GetHeader("Authorization")
	if apiKey == "" {
		apiKey, err = c.Cookie("apikey")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing apikey"})
			return
		}
	}
	userID = c.GetHeader("X-UserID")
	if userID == "" {
		userID, err = c.Cookie("userid")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing user id"})
			return
		}
	}

	userApiKey, err := h.repo.GetApiKeyByUserId(c, userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid apikey"})
		return
	}
	if valid, err := h.hashHelper.Verify(apiKey, userApiKey); err != nil || !valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid apikey"})
		return
	}
	user, err := h.repo.GetUserByID(c, userID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Error: Failed to get user information"})
		return
	}
	c.Set("user", user)
	c.Next()
}
