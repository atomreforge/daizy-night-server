package user

import "github.com/atomreforge/daizy-night-server/internal/model"

type RequestGetNotificationUpdateBatch struct {
	ClientLatestNid uint `json:"client_latest_nid"`
}

type RequestGetNotificationBatch struct {
	NidBegin uint `json:"nid_begin"`
	NidEnd   uint `json:"nid_end"`
}

type RequestGetNotificationSingle struct {
	Nid uint `json:"nid"`
}

type ResponseGetNotificationSingle struct {
	Item      model.Notification `json:"item"`
	NidLatest uint               `json:"nid_latest"`
}

type ResponseGetNotificationBatch struct {
	Batch     []model.Notification `json:"notification_items"`
	NidLatest uint                 `json:"nid_latest"`
}
