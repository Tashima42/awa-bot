package api

type RegisterWaterInput struct {
	Amount             *int  `json:"amount" binding:"required"`
	SendNotification   *bool `json:"sendNotification" binding:"required"`
	NotificationChatID int   `json:"notificationChatID",omitempty`
}

type RegisterWaterOutput struct {
	Success bool `json:"success"`
}
