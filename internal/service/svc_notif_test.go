package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/atomreforge/daizy-night-server/internal/dbware"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	"github.com/atomreforge/daizy-night-server/internal/model"
)

// newNotifSvcEnv spins up an isolated in-memory database through the real
// provider (full schema migrated) and a ServiceNotification bound to it.
func newNotifSvcEnv(t *testing.T) *ServiceNotification {
	t.Helper()

	dsn := fmt.Sprintf("file:dntest_svcnotif_%d?mode=memory&cache=shared", time.Now().UnixNano())
	pDB, err := dbware.NewProviderDB(struct {
		IsDebugMode bool
		DSN         string
	}{IsDebugMode: false, DSN: dsn})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	return &ServiceNotification{RepoNotification: dbware.NewRepoNotification(pDB)}
}

// seedNotifs creates `count` notifications and returns their nids in
// creation order (monotonic nid allocation makes that 1..count).
func seedNotifs(t *testing.T, svc *ServiceNotification, count int) []uint {
	t.Helper()

	nids := make([]uint, 0, count)
	for i := 0; i < count; i++ {
		title := fmt.Sprintf("notif-%02d", i)
		n := model.Notification{
			Md:           model.Markdown{Title: title, Content: "content"},
			TitlePreview: title,
		}
		if err := svc.CreateNotification(n); err != nil {
			t.Fatalf("seed %q: %v", title, err)
		}
		nid, err := svc.GetLatestNotificationNid()
		if err != nil {
			t.Fatalf("latest nid after seed %q: %v", title, err)
		}
		nids = append(nids, nid)
	}
	return nids
}

// A client holding nothing (client_latest_nid == 0) gets the full history;
// an empty repo reports the up-to-date business reply.
func TestBatchUpdateFromZeroAndEmptyRepo(t *testing.T) {
	svc := newNotifSvcEnv(t)

	// empty repo: nothing to sync
	_, err := svc.GetNotificationBatchUpdate(0)
	unnec, ok := errs.Easx[*errs.ErrUnnecBiz](err)
	if !ok {
		t.Fatalf("empty repo update expected *ErrUnnecBiz, got %v", err)
	}
	if unnec.StatusCode() != 400 {
		t.Fatalf("unnec status = %d, want 400", unnec.StatusCode())
	}

	seedNotifs(t, svc, 3)

	batch, err := svc.GetNotificationBatchUpdate(0)
	if err != nil {
		t.Fatalf("full sync: %v", err)
	}
	if len(*batch) != 3 {
		t.Fatalf("full sync len = %d, want 3", len(*batch))
	}
	for i, n := range *batch {
		if n.Nid != uint(i+1) {
			t.Fatalf("full sync[%d].nid = %d, want %d", i, n.Nid, i+1)
		}
	}
}

// The delta is strictly after the client's nid (the already-held item is not
// re-sent).
func TestBatchUpdateIsStrictlyIncremental(t *testing.T) {
	svc := newNotifSvcEnv(t)
	nids := seedNotifs(t, svc, 3)

	batch, err := svc.GetNotificationBatchUpdate(nids[0])
	if err != nil {
		t.Fatalf("incremental sync: %v", err)
	}
	if len(*batch) != 2 {
		t.Fatalf("delta len = %d, want 2", len(*batch))
	}
	if (*batch)[0].Nid != nids[1] || (*batch)[1].Nid != nids[2] {
		t.Fatalf("delta nids = [%d %d], want [%d %d]",
			(*batch)[0].Nid, (*batch)[1].Nid, nids[1], nids[2])
	}
}

// A client already at the latest nid gets the up-to-date business reply.
func TestBatchUpdateAlreadyLatestIsUnnecBiz(t *testing.T) {
	svc := newNotifSvcEnv(t)
	nids := seedNotifs(t, svc, 3)

	_, err := svc.GetNotificationBatchUpdate(nids[2])
	unnec, ok := errs.Easx[*errs.ErrUnnecBiz](err)
	if !ok {
		t.Fatalf("up-to-date update expected *ErrUnnecBiz, got %v", err)
	}
	if unnec.StatusCode() != 400 {
		t.Fatalf("unnec status = %d, want 400", unnec.StatusCode())
	}
}

// Range validation rejects zero bounds and inverted/equal ranges with a
// 400-classified validation error (never a 500).
func TestBatchGetRangeValidation(t *testing.T) {
	svc := newNotifSvcEnv(t)
	seedNotifs(t, svc, 2)

	for _, c := range []struct{ begin, end uint }{
		{0, 2}, {2, 0}, {0, 0}, {2, 2}, {3, 2},
	} {
		_, err := svc.GetNotificationBatchByNid(c.begin, c.end)
		valErr, ok := errs.Easx[*errs.ErrValidation](err)
		if !ok {
			t.Fatalf("range (%d,%d): expected *ErrValidation, got %v", c.begin, c.end, err)
		}
		if valErr.StatusCode() != 400 {
			t.Fatalf("range (%d,%d): status = %d, want 400", c.begin, c.end, valErr.StatusCode())
		}
	}
}

// A nid beyond the latest one is a client-side bad value (400).
func TestGetSingleRejectsNidBeyondLatest(t *testing.T) {
	svc := newNotifSvcEnv(t)
	seedNotifs(t, svc, 3)

	_, err := svc.GetNotificationSingleByNid(99)
	valErr, ok := errs.Easx[*errs.ErrValidation](err)
	if !ok {
		t.Fatalf("expected *ErrValidation, got %v", err)
	}
	if valErr.StatusCode() != 400 {
		t.Fatalf("status = %d, want 400", valErr.StatusCode())
	}
}

// A single get of a removed notification maps to 404. The removed nid must
// sit inside the live range (nid <= latest) so the request reaches the
// lookup; beyond-latest nids are rejected as bad values before the query.
func TestGetSingleRemovedNidIs404(t *testing.T) {
	svc := newNotifSvcEnv(t)
	nids := seedNotifs(t, svc, 2)

	if err := svc.RemoveNotification(nids[0]); err != nil {
		t.Fatalf("remove: %v", err)
	}
	_, err := svc.GetNotificationSingleByNid(nids[0])
	recErr, ok := errs.Easx[*errs.ErrDbRecord](err)
	if !ok {
		t.Fatalf("expected *ErrDbRecord, got %v", err)
	}
	if recErr.StatusCode() != 404 {
		t.Fatalf("status = %d, want 404", recErr.StatusCode())
	}
}

// Post stamps post_by/post_at through the service layer; withdraw clears
// them again while the notification itself stays available.
func TestPostAndWithdrawLifecycle(t *testing.T) {
	svc := newNotifSvcEnv(t)
	nids := seedNotifs(t, svc, 2)

	if err := svc.PostNotification(nids[0], 42); err != nil {
		t.Fatalf("post: %v", err)
	}
	posted, err := svc.GetNotificationSingleByNid(nids[0])
	if err != nil {
		t.Fatalf("get posted: %v", err)
	}
	if posted.PostBy != 42 || posted.PostAt == 0 {
		t.Fatalf("post stamp missing: post_by=%d post_at=%d", posted.PostBy, posted.PostAt)
	}

	if err := svc.WithdrawNotification(nids[0]); err != nil {
		t.Fatalf("withdraw: %v", err)
	}
	withdrawn, err := svc.GetNotificationSingleByNid(nids[0])
	if err != nil {
		t.Fatalf("get withdrawn: %v", err)
	}
	if withdrawn.PostBy != 0 || withdrawn.PostAt != 0 {
		t.Fatalf("withdraw did not clear stamp: post_by=%d post_at=%d", withdrawn.PostBy, withdrawn.PostAt)
	}
}

// Delete soft-deletes: the nid 404s afterwards.
func TestDeleteNotificationSoftDeletes(t *testing.T) {
	svc := newNotifSvcEnv(t)
	nids := seedNotifs(t, svc, 2)

	if err := svc.DeleteNotification(nids[0]); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err := svc.GetNotificationSingleByNid(nids[0])
	recErr, ok := errs.Easx[*errs.ErrDbRecord](err)
	if !ok || recErr.StatusCode() != 404 {
		t.Fatalf("expected 404 ErrDbRecord after delete, got %v", err)
	}
}