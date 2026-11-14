package dbware

import (
	"fmt"
	"testing"
	"time"

	"github.com/atomreforge/daizy-night-server/internal/config"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	"github.com/atomreforge/daizy-night-server/internal/model"
	"github.com/atomreforge/daizy-night-server/internal/utils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newTokenTestEnv spins up an isolated in-memory database with the refresh
// token table migrated and a RepoToken bound to it (168h lifetime, 72h
// retention, the production defaults).
func newTokenTestEnv(t *testing.T) *RepoToken {
	t.Helper()

	dsn := fmt.Sprintf("file:dntest_token_%d?mode=memory&cache=shared", time.Now().UnixNano())
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := gdb.AutoMigrate(&model.RefreshToken{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cfg := &config.Config{}
	cfg.Security.JwtRefreshTokenExpireTime = 168 * time.Hour
	cfg.Security.JwtRevokedTokensRetainTime = 72 * time.Hour

	return &RepoToken{pDB: &ProviderDB{db: gdb}, cfg: cfg}
}

func mustInsertToken(t *testing.T, r *RepoToken, created, revoked *time.Time, lookup string) {
	t.Helper()
	rt := model.RefreshToken{
		Model:      gorm.Model{CreatedAt: *created, UpdatedAt: *created},
		Uid:        1,
		LookupHash: lookup,
	}
	if revoked != nil {
		rt.RevokedAt = revoked
	}
	if err := r.pDB.DB().Create(&rt).Error; err != nil {
		t.Fatalf("seed token %s: %v", lookup, err)
	}
}

func countTokens(t *testing.T, r *RepoToken, lookup string) int64 {
	t.Helper()
	var n int64
	if err := r.pDB.DB().Model(&model.RefreshToken{}).
		Where("lookup_hash = ?", lookup).Count(&n).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	return n
}

// A naturally-expired token (revoked_at IS NULL) must be pruned once it is
// older than refresh lifetime + retention; younger ones and active tokens
// must survive.
func TestPruneStaleTokensRemovesExpired(t *testing.T) {
	r := newTokenTestEnv(t)
	now := time.Now()

	expired := now.Add(-241 * time.Hour)      // strictly beyond 168h lifetime + 72h buffer: gone
	withinBuffer := now.Add(-180 * time.Hour) // expired, but inside the buffer: kept
	active := now.Add(-2 * time.Hour)         // alive: kept
	mustInsertToken(t, r, &expired, nil, "expired")
	mustInsertToken(t, r, &withinBuffer, nil, "buffer")
	mustInsertToken(t, r, &active, nil, "active")

	if err := r.PruneStaleTokens(1); err != nil {
		t.Fatalf("prune: %v", err)
	}

	if got := countTokens(t, r, "expired"); got != 0 {
		t.Fatalf("expired token not pruned, count = %d", got)
	}
	if got := countTokens(t, r, "buffer"); got != 1 {
		t.Fatalf("token within retention buffer was pruned, count = %d", got)
	}
	if got := countTokens(t, r, "active"); got != 1 {
		t.Fatalf("active token was pruned, count = %d", got)
	}
}

// Revoked tokens are kept for the retention window after revocation, then
// removed, independently of their creation time.
func TestPruneStaleTokensRemovesOldRevoked(t *testing.T) {
	r := newTokenTestEnv(t)
	now := time.Now()

	created := now.Add(-400 * time.Hour)
	oldRevoked := now.Add(-73 * time.Hour) // past the 72h retention: gone
	freshRevoked := now.Add(-1 * time.Hour)

	mustInsertToken(t, r, &created, &oldRevoked, "old-revoked")
	mustInsertToken(t, r, &created, &freshRevoked, "fresh-revoked")

	if err := r.PruneStaleTokens(1); err != nil {
		t.Fatalf("prune: %v", err)
	}

	if got := countTokens(t, r, "old-revoked"); got != 0 {
		t.Fatalf("old revoked token not pruned, count = %d", got)
	}
	if got := countTokens(t, r, "fresh-revoked"); got != 1 {
		t.Fatalf("freshly revoked token was pruned, count = %d", got)
	}
}

// Rotation revokes the used token atomically with inserting the new one; the
// used token can never be rotated again.
func TestRotateRefreshTokenRevokesUsedToken(t *testing.T) {
	r := newTokenTestEnv(t)

	if err := r.SaveRefreshToken(1, "token-one"); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := r.RotateRefreshToken(1, utils.SHA256HashHex("token-one"), "token-two"); err != nil {
		t.Fatalf("rotate: %v", err)
	}

	var used model.RefreshToken
	if err := r.pDB.DB().Where("lookup_hash = ?", utils.SHA256HashHex("token-one")).First(&used).Error; err != nil {
		t.Fatalf("load used token: %v", err)
	}
	if used.RevokedAt == nil {
		t.Fatal("used token was not revoked by rotation")
	}
	if got := countTokens(t, r, utils.SHA256HashHex("token-two")); got != 1 {
		t.Fatalf("replacement token missing, count = %d", got)
	}

	// replaying the already-used token must fail with 401 semantics
	err := r.RotateRefreshToken(1, utils.SHA256HashHex("token-one"), "token-three")
	if err == nil {
		t.Fatal("replay of used token unexpectedly succeeded")
	}
	if errapp, ok := errs.Easx[*errs.ErrJwtToken](err); !ok || errapp.StatusCode() != 401 {
		t.Fatalf("expected 401 ErrJwtToken on replay, got %v", err)
	}
}
