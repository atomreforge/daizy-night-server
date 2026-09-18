package dbware

import (
	"fmt"
	"testing"
	"time"

	"github.com/atomreforge/daizy-night-server/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// newCalendarTestEnv spins up an isolated in-memory database with the
// calendar schema migrated and a RepoCalendar bound to it.
func newCalendarTestEnv(t *testing.T) *RepoCalendar {
	t.Helper()

	dsn := fmt.Sprintf("file:dntest_calendar_%d?mode=memory&cache=shared", time.Now().UnixNano())
	gdb, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := gdb.AutoMigrate(&model.CalendarTable{}, &model.CalendarItem{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return &RepoCalendar{pDB: &ProviderDB{db: gdb}}
}

func sampleCalendarRecords() []model.CalendarItem {
	return []model.CalendarItem{
		{Weekday: 1, StartMin: 480, EndMin: 540, Title: "CS"},
		{Weekday: 3, StartMin: 600, EndMin: 660, Title: "Math"},
	}
}

// PUT is a full replacement: a second UpdateCalendar must swap the items AND
// persist the calendar-level roaming columns (the update path previously
// never touched the calendar row itself).
func TestUpdateCalendarReplacesRoamingAndItems(t *testing.T) {
	r := newCalendarTestEnv(t)

	first := &model.CalendarTable{
		UserID:  42,
		Roaming: model.Roaming{Description: "first-desc", Annotation: "first-note"},
		Records: sampleCalendarRecords(),
	}
	if err := r.UpdateCalendar(first); err != nil {
		t.Fatalf("first update (create path): %v", err)
	}

	second := &model.CalendarTable{
		UserID:  42,
		Roaming: model.Roaming{Description: "second-desc", Annotation: "second-note"},
		Records: []model.CalendarItem{
			{Weekday: 2, StartMin: 100, EndMin: 200, Title: "Physics", Roaming: model.Roaming{Description: "lab"}},
		},
	}
	if err := r.UpdateCalendar(second); err != nil {
		t.Fatalf("second update: %v", err)
	}

	got, err := r.GetCalendarByUid(42)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.Description != "second-desc" || got.Annotation != "second-note" {
		t.Fatalf("calendar-level roaming not persisted on update: %+v", got.Roaming)
	}
	if len(got.Records) != 1 {
		t.Fatalf("items not swapped: got %d records", len(got.Records))
	}
	if got.Records[0].Title != "Physics" || got.Records[0].Description != "lab" {
		t.Fatalf("item roaming not persisted: %+v", got.Records[0])
	}
}

// An omitted (zero) roaming in the replacement body must clear the stored
// values instead of being skipped by gorm's zero-value struct updates.
func TestUpdateCalendarZeroRoamingClearsStoredValues(t *testing.T) {
	r := newCalendarTestEnv(t)

	first := &model.CalendarTable{
		UserID:  43,
		Roaming: model.Roaming{Description: "d", Annotation: "a"},
		Records: sampleCalendarRecords(),
	}
	if err := r.UpdateCalendar(first); err != nil {
		t.Fatalf("first update: %v", err)
	}

	if err := r.UpdateCalendar(&model.CalendarTable{UserID: 43, Records: sampleCalendarRecords()}); err != nil {
		t.Fatalf("second update: %v", err)
	}

	got, err := r.GetCalendarByUid(43)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.Description != "" || got.Annotation != "" {
		t.Fatalf("expected roaming cleared, got %+v", got.Roaming)
	}
	if len(got.Records) != 2 {
		t.Fatalf("expected 2 records after re-put, got %d", len(got.Records))
	}
}
