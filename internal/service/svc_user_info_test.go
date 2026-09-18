package service

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/atomreforge/daizy-night-server/internal/dbware"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	"github.com/atomreforge/daizy-night-server/internal/model"
)

// newInfoSvcEnv spins up an isolated in-memory database through the real
// provider (full schema migrated) and a ServiceUser bound to it. Only the
// user repo is exercised by the /info service methods, so the remaining
// collaborators stay nil.
func newInfoSvcEnv(t *testing.T) (*ServiceUser, *dbware.ProviderDB) {
	t.Helper()

	dsn := fmt.Sprintf("file:dntest_svcinfo_%d?mode=memory&cache=shared", time.Now().UnixNano())
	pDB, err := dbware.NewProviderDB(struct {
		IsDebugMode bool
		DSN         string
	}{IsDebugMode: false, DSN: dsn})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	return NewServiceUser(dbware.NewRepoUser(pDB), nil, nil, nil, nil), pDB
}

func TestGetInfoByUsernameMapsFullView(t *testing.T) {
	svc, pDB := newInfoSvcEnv(t)

	regTime := time.Now()
	u := &model.User{
		UserID:       1527277,
		Username:     "alice",
		Nickname:     "Alice",
		PasswordHash: "hash",
		Registercode: "aa.bb",
		Email:        "alice@example.com",
		RegisterTime: regTime,
	}
	if err := pDB.DB().Create(u).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	info, err := svc.GetInfoByUsername("alice")
	if err != nil {
		t.Fatalf("GetInfoByUsername: %v", err)
	}
	if info.Uid != u.UserID || info.Username != "alice" || info.Nickname != "Alice" ||
		info.Email != u.Email || info.RegisterTime.IsZero() || info.Role != u.Role {
		t.Fatalf("unexpected mapping: %+v", info)
	}
	if info.GithubID != nil || info.GithubLogin != nil {
		t.Fatalf("github fields must pass through as-is: %+v", info)
	}
}

func TestGetInfoByUsernameUnknownUserIsRemapped404(t *testing.T) {
	svc, _ := newInfoSvcEnv(t)

	_, err := svc.GetInfoByUsername("ghost")
	if err == nil {
		t.Fatal("expected error for unknown username")
	}

	e, ok := errs.Easx[*errs.ErrDbRecord](err)
	if !ok {
		t.Fatalf("expected *errs.ErrDbRecord, got %T", err)
	}
	if e.Type != errs.DbRecordUsernameNotFound {
		t.Fatalf("expected DbRecordUsernameNotFound, got %v", e.Type)
	}
	if code := e.StatusCode(); code != http.StatusNotFound {
		t.Fatalf("expected remapped 404, got %d", code)
	}
}
