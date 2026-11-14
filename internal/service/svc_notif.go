package service

import (
	"fmt"
	"net/http"
	"strconv"

	abstract "github.com/atomreforge/daizy-night-server/internal/abstract/interface"
	"github.com/atomreforge/daizy-night-server/internal/consts"
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

func (s *ServiceNotification) CreateNotification(n model.Notification) error {
	return s.RepoNotification.CreateItem(n)
}

func (s *ServiceNotification) RemoveNotification(nid uint) error {
	return s.RepoNotification.RemoveItem(nid)
}

func (s *ServiceNotification) UpdateNotification(n model.Notification) error {
	return s.RepoNotification.UpdateItem(n)
}

// havent ok
func (s *ServiceNotification) PostNotification(n *model.Notification) error {
	// remb to override attrs `PostBy` and `PostAt`

	// unsupported feature
	return errs.BuildErrSupport(
		errs.FeatureUnsupported,
		http.StatusInternalServerError,
	)
}

// havent ok
func (s *ServiceNotification) WithdrawNotification(nid uint) error {
	// unsupported feature
	return errs.BuildErrSupport(
		errs.FeatureUnsupported,
		http.StatusInternalServerError,
	)
}

// havent ok
func (s *ServiceNotification) DeleteNotification(nid uint) error {
	// unsupported feature
	return errs.BuildErrSupport(
		errs.FeatureUnsupported,
		http.StatusInternalServerError,
	)
}

func (s *ServiceNotification) GetNotificationSingleByNid(nid uint) (*model.Notification, error) {
	// get latest nid from repo
	nidLatest, err := s.GetLatestNotificationNid()
	if err != nil {
		return nil, err
	}

	// compare nid & nidLatest(in repo)
	if nid > nidLatest {
		return nil, errs.BuildErrValidation(
			errs.ValidationKeyBadValue,
			http.StatusBadRequest,
			string(consts.JsonExprNid),
			strconv.FormatUint(uint64(nid), 10),
		)
	}

	// formal query
	n, err := s.RepoNotification.GetItemByNid(nid)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (s *ServiceNotification) GetNotificationBatchByNid(nidBegin uint, nidEnd uint) (*[]model.Notification, error) {
	// validate; nid is a strictly positive, monotonic integer, so zero is
	// never a valid bound and begin must stay strictly below end
	if nidBegin == 0 || nidEnd == 0 || nidBegin >= nidEnd {
		return nil, errs.BuildErrValidation(
			errs.ValidationKeyBadRange,
			http.StatusBadRequest,
			string(consts.UiExprNid),
			fmt.Sprintf("%d, %d", nidBegin, nidEnd),
		)
	}

	// formal query
	batch, err := s.RepoNotification.GetNotificationBatch(nidBegin, nidEnd)
	if err != nil {
		return nil, err
	}

	// return
	return batch, nil
}

// get latest nid
func (s *ServiceNotification) GetLatestNotificationNid() (uint, error) {
	latestNid, err := s.RepoNotification.GetLatestNotificationNid()
	if err != nil {
		return 0, err
	}
	return latestNid, nil
}

// GetNotificationBatchUpdate returns every notification strictly after the
// client's `clientLatestNid`, up to the server's latest one. With monotonic
// nid allocation the nid ordering equals the creation ordering, so a numeric
// window is a faithful delta. `clientLatestNid == 0` means the client holds
// nothing and yields the full history; a client that is already up-to-date
// (or ahead, e.g. after server-side removals) gets a 400 UnnecessaryBiz
// reply instead of an empty payload.
func (s *ServiceNotification) GetNotificationBatchUpdate(clientLatestNid uint) (*[]model.Notification, error) {
	// server-side latest nid; 0 when the repo is empty
	repoLatestNid, err := s.RepoNotification.GetLatestNotificationNid()
	if err != nil {
		return nil, err
	}

	// nothing newer exists for the client
	if clientLatestNid >= repoLatestNid {
		return nil, errs.BuildErrUnnecBiz(
			errs.UnnecBizNidAlreadyLatest,
			http.StatusBadRequest,
			string(consts.UiExprNid),
		)
	}

	// strictly after the client's nid
	batch, err := s.RepoNotification.GetNotificationBatch(clientLatestNid+1, repoLatestNid)
	if err != nil {
		return nil, err
	}
	return batch, nil
}
