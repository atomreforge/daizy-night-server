package dbware

import (
	"log/slog"
	"net/http"

	abstract "github.com/atomreforge/daizy-night-server/internal/abstract/interface"
	"github.com/atomreforge/daizy-night-server/internal/errs"
	"github.com/atomreforge/daizy-night-server/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _ abstract.InterfaceProviderDB = (*ProviderDB)(nil)

type ProviderDB struct {
	db *gorm.DB
}

func NewProviderDB(ctx struct {
	IsDebugMode bool
	DSN         string
}) (*ProviderDB, error) {
	logLevel := logger.Warn
	if ctx.IsDebugMode {
		logLevel = logger.Info
	}

	gormLogger := logger.NewSlogLogger(slog.Default(), logger.Config{
		LogLevel: logLevel,
	})

	// create or connect db
	db, err := gorm.Open(sqlite.Open(ctx.DSN), &gorm.Config{
		Logger:         gormLogger,
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.RefreshToken{},
		&model.RegistercodeRecord{},
		&model.CalendarTable{}, // AutoMigrate also creates the associated calendar_items table
	); err != nil {
		return nil, err
	}

	// legacy migration for databases created before the TokenHash removal
	if err := migrateLegacyTokenHash(db); err != nil {
		return nil, err
	}

	return &ProviderDB{db: db}, nil
}

// migrateLegacyTokenHash rebuilds refresh_tokens without the removed
// token_hash column. Legacy tables declare it through a table-level UNIQUE
// constraint (the gorm `unique` tag compiles to CONSTRAINT ... UNIQUE),
// which neither ALTER TABLE DROP COLUMN nor Migrator.DropColumn can remove —
// the constraint clause keeps referencing the dropped column — so the table
// is renamed, recreated fresh, and its rows are copied over explicitly.
func migrateLegacyTokenHash(db *gorm.DB) error {
	var hasTokenHash int64
	if err := db.Raw(`SELECT count(*) FROM pragma_table_info('refresh_tokens') WHERE name = 'token_hash'`).Scan(&hasTokenHash).Error; err != nil {
		return err
	}
	if hasTokenHash == 0 {
		return nil
	}

	const legacy = "refresh_tokens_legacy"
	if err := db.Migrator().RenameTable("refresh_tokens", legacy); err != nil {
		return err
	}

	// named indexes travel with the renamed table and would collide with the
	// ones AutoMigrate creates on the fresh table; drop them up front. the
	// autoindexes of table-level constraints go away with DROP TABLE.
	var idxNames []string
	if err := db.Raw(`SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = ? AND name NOT LIKE 'sqlite_autoindex_%'`, legacy).Scan(&idxNames).Error; err != nil {
		return err
	}
	for _, idx := range idxNames {
		if err := db.Exec("DROP INDEX IF EXISTS " + idx).Error; err != nil {
			return err
		}
	}

	if err := db.AutoMigrate(&model.RefreshToken{}); err != nil {
		return err
	}
	if err := db.Exec(`INSERT INTO refresh_tokens
		(id, created_at, updated_at, deleted_at, uid, lookup_hash, revoked_at)
		SELECT id, created_at, updated_at, deleted_at, uid, lookup_hash, revoked_at
		FROM ` + legacy).Error; err != nil {
		return err
	}
	return db.Migrator().DropTable(legacy)
}

// check db health
func (p *ProviderDB) Check() error {
	if p.db == nil {
		return errs.BuildErrConnDb(
			errs.ConnDb,
			http.StatusInternalServerError,
		)
	}
	sql, err := p.db.DB()
	if err != nil {
		return err
	}
	return sql.Ping()
}

func (p *ProviderDB) Close() error {
	sql, err := p.db.DB()
	if err != nil {
		return err
	}
	return sql.Close()
}

func (p *ProviderDB) DB() *gorm.DB { return p.db }
