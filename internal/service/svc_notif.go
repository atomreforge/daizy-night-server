package service

import (
	"net/http"

	abstract "github.com/atomreforge/daizy-night-server/internal/abstract/interface"
	"github.com/atomreforge/daizy-night-server/internal/dbware"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	"github.com/atomreforge/daizy-night-server/internal/model"
)

var _ abstract.InterfaceServiceNotification = (*ServiceNotification)(nil)

type ServiceNotification struct {
	RepoNotification *dbware.RepoNotification
}

func NewServiceNotification(repoNotification *dbware.RepoNotification) *ServiceNotification {
	return &ServiceNotification{
		RepoNotification: repoNotification,
	}
}

func CreateServiceNotification(repoNotification *dbware.RepoNotification) *ServiceNotification {
	return &ServiceNotification{
		RepoNotification: repoNotification,
	}
}

func (s *ServiceNotification) CreateNotification(n model.Notification) error {
	return s.RepoNotification.CreateItem(n)
}

func (s *ServiceNotification) RemoveNotification(nid uint) error {
	return s.RepoNotification.RemoveItem(nid)
}

func (s *ServiceNotification) GetNotificationSingelByNid(nid uint) (*model.Notification, error) {
	n, err := s.RepoNotification.GetItemByNid(nid)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (s *ServiceNotification) GetNotificationBatchByNid(nidBegin uint, nidEnd uint) (*[]model.Notification, error) {
	if nidBegin >= nidEnd {
		return nil, errs.BuildErrInternalServiceCall(errs.ValidationKeyBadValue)
	}
	var batch []model.Notification
	for nid := nidBegin; nid <= nidEnd; nid++ {
		n, err := s.RepoNotification.GetItemByNid(nid)
		if err != nil {
			return nil, err
		}
		batch = append(batch, n)
	}
	// Implementation for fetching batch notifications by Nid
	return &batch, nil
}

func (s *ServiceNotification) UpdateNotification(n model.Notification) error {
	return s.RepoNotification.UpdateItem(n)
}

// havent ok
func (s *ServiceNotification) PostNotification(n *model.Notification) error {
	return errs.BuildErrSupport(
		errs.FeatureUnsupported,
		http.StatusInternalServerError,
	)
}

// get latest nid
func (s *ServiceNotification) GetLatestNotificationNid() (uint, error) {
	latestNid, err := s.RepoNotification.GetLatestNotificationNid()
	if err != nil {
		return 0, err
	}
	return latestNid, nil
}
