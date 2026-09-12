package v1

import "github.com/atomreforge/daizy-night-server/internal/model"

type InfoMeRequest struct {
}

type InfoMeResponse struct {
	model.InfoUser
}
