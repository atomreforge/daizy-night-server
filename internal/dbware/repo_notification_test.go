package dbware

import (
	"fmt"
	"testing"
	"time"

	"github.com/atomreforge/daizy-night-server/internal/errs"
	"github.com/atomreforge/daizy-night-server/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newNotifTestEnv spins up an isolated in-memory database with the
// notification table migrated and a RepoNotification bound to it.
func newNotifTestEnv(t *testing.T) *RepoNotification {
	t.Helper()

	dsn := fmt.Sprintf("file:dntest_notif_%d?mode=memory&cache=shared", time.Now().UnixNano())
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := gdb.AutoMigrate(&model.Notification{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	return &RepoNotification{pDB: &ProviderDB{db: gdb}}
}

// mustCreateNotif inserts a notification and returns its server-assigned nid.
func mustCreateNotif(t *testing.T, r *RepoNotification, title string) uint {
	t.Helper()
	n := model.Notification{
		Md:           model.Markdown{Title: title, Content: "content of " + title},
		TitlePreview: title,
	}
	if err := r.CreateItem(n); err != nil {
		t.Fatalf("create %q: %v", title, err)
	}
	nid, err := r.GetLatestNotificationNid()
	if err != nil {
		t.Fatalf("latest nid after create %q: %v", title, err)
	}
	return nid
}

// Nid is a strictly positive, monotonically increasing server-assigned
// integer: 1, 2, 3, ... regardless of what the caller passes in.
func TestCreateItemAllocatesSequentialPositiveNid(t *testing.T) {
	r := newNotifTestEnv(t)

	if nid := mustCreateNotif(t, r, "first"); nid != 1 {
		t.Fatalf("first nid = %d, want 1", nid)
	}
	if nid := mustCreateNotif(t, r, "second"); nid != 2 {
		t.Fatalf("second nid = %d, want 2", nid)
	}
	if nid := mustCreateNotif(t, r, "third"); nid != 3 {
		t.Fatalf("third nid = %d, want 3", nid)
	}
}

// A caller-provided Nid is ignored: the repo never persists client-chosen
// ids (they could collide with the unique index or break the monotonic order).
func TestCreateItemIgnoresCallerProvidedNid(t *testing.T) {
	r := newNotifTestEnv(t)

	n := model.Notification{
		Md:           model.Markdown{Title: "forged", Content: "forged content"},
		TitlePreview: "forged",
		Nid:          9999,
	}
	if err := r.CreateItem(n); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := r.GetItemByNid(9999); err == nil {
		t.Fatal("caller-provided nid 9999 must not be persisted")
	}
	if nid, err := r.GetLatestNotificationNid(); err != nil || nid != 1 {
		t.Fatalf("allocated nid = %d (err %v), want 1", nid, err)
	}
}

// Nids of soft-deleted notifications are never reused: the allocation
// counter runs over every row, deleted ones included.
func TestCreateItemNeverReusesSoftDeletedNid(t *testing.T) {
	r := newNotifTestEnv(t)

	first := mustCreateNotif(t, r, "doomed")
	if err := r.RemoveItem(first); err != nil {
		t.Fatalf("remove: %v", err)
	}

	nid := mustCreateNotif(t, r, "after-delete")
	if nid != first+1 {
		t.Fatalf("nid after removal = %d, want %d (no reuse)", nid, first+1)
	}
	_, err := r.GetItemByNid(first)
	if !errs.Eas[*errs.ErrDbRecord](err) {
		t.Fatalf("removed nid must stay unavailable, err = %v", err)
	}
}

// GetLatestNotificationNid: 0 on an empty repo; the greatest live nid
// otherwise (soft-deleted rows do not count).
func TestGetLatestNotificationNid(t *testing.T) {
	r := newNotifTestEnv(t)

	if nid, err := r.GetLatestNotificationNid(); err != nil || nid != 0 {
		t.Fatalf("empty repo latest = %d (err %v), want 0", nid, err)
	}

	mustCreateNotif(t, r, "one")
	mustCreateNotif(t, r, "two")
	third := mustCreateNotif(t, r, "three")

	if err := r.RemoveItem(third); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if nid, err := r.GetLatestNotificationNid(); err != nil || nid != third-1 {
		t.Fatalf("latest after removing the max = %d (err %v), want %d", nid, err, third-1)
	}
}

// GetItemByNid maps a missing record to a 404-classified ErrDbRecord.
func TestGetItemByNidNotFoundIs404(t *testing.T) {
	r := newNotifTestEnv(t)
	mustCreateNotif(t, r, "only")

	_, err := r.GetItemByNid(42)
	if err == nil {
		t.Fatal("expected error for missing nid")
	}
	recErr, ok := errs.Easx[*errs.ErrDbRecord](err)
	if !ok {
		t.Fatalf("expected *errs.ErrDbRecord, got %T", err)
	}
	if recErr.StatusCode() != 404 {
		t.Fatalf("missing nid status = %d, want 404", recErr.StatusCode())
	}
}

// GetNotificationBatch returns the closed nid range ascending and never a
// nil slice (an empty range marshals to []).
func TestGetNotificationBatchRangeAndOrdering(t *testing.T) {
	r := newNotifTestEnv(t)

	for _, title := range []string{"a", "b", "c", "d"} {
		mustCreateNotif(t, r, title)
	}

	batch, err := r.GetNotificationBatch(2, 3)
	if err != nil {
		t.Fatalf("batch: %v", err)
	}
	if len(*batch) != 2 || (*batch)[0].Nid != 2 || (*batch)[1].Nid != 3 {
		t.Fatalf("batch(2,3) = %+v, want nids [2 3]", *batch)
	}

	empty, err := r.GetNotificationBatch(9, 10)
	if err != nil {
		t.Fatalf("empty batch: %v", err)
	}
	if empty == nil || *empty == nil {
		t.Fatal("empty batch must not be a nil slice")
	}
	if len(*empty) != 0 {
		t.Fatalf("empty batch len = %d, want 0", len(*empty))
	}
}

// UpdateItem replaces the content fields but keeps server-owned columns
// (created_at / deleted_at / post_by) intact.
func TestUpdateItemPreservesServerOwnedColumns(t *testing.T) {
	r := newNotifTestEnv(t)

	nid := mustCreateNotif(t, r, "original")
	rec, err := r.GetItemByNid(nid)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	origCreated := rec.CreatedAt
	origPostBy := rec.PostBy

	updated := model.Notification{
		Md:           model.Markdown{Title: "edited", Content: "edited content"},
		Nid:          nid,
		TitlePreview: "edited",
	}
	if err := r.UpdateItem(updated); err != nil {
		t.Fatalf("update: %v", err)
	}

	got, err := r.GetItemByNid(nid)
	if err != nil {
		t.Fatalf("re-get: %v", err)
	}
	if got.TitlePreview != "edited" || got.Md.Content != "edited content" {
		t.Fatalf("content not updated: %+v", got)
	}
	if !got.CreatedAt.Equal(origCreated) {
		t.Fatalf("created_at changed: %v -> %v", origCreated, got.CreatedAt)
	}
	if got.PostBy != origPostBy {
		t.Fatalf("post_by changed: %d -> %d", origPostBy, got.PostBy)
	}
}