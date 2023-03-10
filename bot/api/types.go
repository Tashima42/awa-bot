package api

import "github.com/tashima42/awa-bot/bot/pkg/db"

type RegisterWaterInput struct {
	Amount             *int  `json:"amount" binding:"required"`
	SendNotification   *bool `json:"sendNotification" binding:"required"`
	NotificationChatID int   `json:"notificationChatID",omitempty`
}

type RegisterWaterOutput struct {
	Success bool `json:"success"`
}

type GetWaterOutput struct {
	Waters []db.Water `json:"waters"`
	Total  int        `json:"total"`
}

type LoginInput struct {
	Code string `json:"code" binding:"required"`
}

type LoginOutput struct {
	Success bool   `json:"success"`
	UserID  string `json:"userID"`
}
