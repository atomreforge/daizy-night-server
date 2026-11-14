package user

import "github.com/atomreforge/daizy-night-server/internal/model"

type ResponseCreateNotification struct {
	Message string `json:"message"`
}

type ResponseUpdateNotification struct {
	Message string `json:"message"`
}

type ResponseDeleteNotification struct {
	Message string `json:"message"`
}

type ResponseWithdrawNotification struct {
	Message string `json:"message"`
}

type ResponsePostNotification struct {
	Message string `json:"message"`
}

// share the same resp body with `ResponseGetNotificationBatch`
type ResponseGetNotificationBatchUpdate struct {
	Payload []model.Notification
	Message string `json:"message"`
}

type ResponseGetNotificationBatch struct {
	Payload []model.Notification
	Message string `json:"message"`
}

type ResponseGetNotificationSingle struct {
	Payload model.Notification
	Message string `json:"message"`
}

type ResponseUploadAndPostNotification struct {
	Message string `json:"message"`
}
