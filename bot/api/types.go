package api

type RegisterWaterInput struct {
	Amount             int  `json:"amount"`
	SendNotification   bool `json:"sendNotification"`
	NotificationChatID int  `json:"notificationChatID",omitempty`
}

type RegisterWaterOutput struct {
	Success bool `json:"success"`
}
