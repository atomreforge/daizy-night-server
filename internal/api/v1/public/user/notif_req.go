package user

import (
	"github.com/atomreforge/daizy-night-server/internal/consts"
	"github.com/atomreforge/daizy-night-server/internal/model"
)

// request

// create a notif without posting it;
// req admin;
type RequestCreateNotification struct {
	Md           model.Markdown
	TitlePreview string
	Levels       consts.Levels
}

// update a existing notif;
// req admin
type RequestUpdateNotification struct {
	Md           model.Markdown
	TitlePreview string
	Levels       consts.Levels
	Nid          uint
}

// delete a notif;
// req admin;
type RequestDeleteNotification struct {
	Nid uint `json:"nid"`
}

// withdraw a posted notif;
// req admin;
type RequestWithdrawNotification struct {
	Nid uint `json:"nid"`
}

// post a notif
type RequestPostNotification struct {
	Nid uint `json:"nid"`
}

// backend compares the latest `Nid` in the repo with that client has, and returns a batch if appropriate.
type RequestGetNotificationBatchUpdate struct {
	ClientLatestNid uint `json:"client_latest_nid"`
}

// request for batch of notification items, with a range of `Nid`.
type RequestGetNotificationBatch struct {
	NidBegin uint `json:"nid_begin"`
	NidEnd   uint `json:"nid_end"`
}

// request for a single notification item, with a specific `Nid`.
type RequestGetNotificationSingle struct {
	Nid uint `json:"nid"`
}
