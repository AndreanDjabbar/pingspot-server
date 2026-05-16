package migration

import (
	"pingspot/internal/model"
	"pingspot/pkg/logger"

	"github.com/go-gormigrate/gormigrate/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "11142025_initial_migration",
			Migrate: func(tx *gorm.DB) error {
				if err := tx.AutoMigrate(
					&model.User{},
					&model.UserProfile{},
					&model.UserSession{},
					&model.Report{},
					&model.ReportLocation{},
					&model.ReportImage{},
					&model.ReportReaction{},
					&model.ReportProgress{},
					&model.ReportVote{},
				); err != nil {
					return err
				}

				if err := tx.Exec(`
					ALTER TABLE reports
					ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS (
						setweight(to_tsvector('indonesian', coalesce(report_title, '')), 'A') ||
						setweight(to_tsvector('indonesian', coalesce(report_description, '')), 'B')
					) STORED;
				`).Error; err != nil {
					return err
				}

				if err := tx.Exec(`
					ALTER TABLE users
					ADD COLUMN IF NOT EXISTS search_vector tsvector GENERATED ALWAYS AS (
						setweight(to_tsvector('indonesian', coalesce(username, '')), 'A') ||
						setweight(to_tsvector('indonesian', coalesce(email, '')), 'B') ||
						setweight(to_tsvector('indonesian', coalesce(full_name, '')), 'C')
					) STORED;
				`).Error; err != nil {
					return err
				}

				if err := tx.Exec(`
					CREATE INDEX IF NOT EXISTS idx_reports_search_vector
					ON reports USING GIN (search_vector);
				`).Error; err != nil {
					return err
				}

				if err := tx.Exec(`
					CREATE INDEX IF NOT EXISTS idx_users_search_vector
					ON users USING GIN (search_vector);
				`).Error; err != nil {
					return err
				}

				return nil
			},

			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable(
					&model.User{},
					&model.UserProfile{},
					&model.UserSession{},
					&model.Report{},
					&model.ReportLocation{},
					&model.ReportImage{},
					&model.ReportReaction{},
					&model.ReportProgress{},
					&model.ReportVote{},
				)
			},
		},
		{
			ID: "04052026_add_is_default_username_to_users",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&model.User{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropColumn(&model.User{}, "is_default_username")
			},
		},
		{
			ID: "04052026_add_last_reminder_at_to_users",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&model.User{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropColumn(&model.User{}, "last_reminder_at")
			},
		},
		{
			ID: "05162026_remove_last_reminder_at_from_users",
			Migrate: func(tx *gorm.DB) error {
				return tx.Migrator().DropColumn(&model.User{}, "last_reminder_at")
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&model.User{})
			},
		},
	})

	err := m.Migrate()
	if err != nil {
		logger.Error("Failed to run migrations", zap.Error(err))
		return err
	}
	logger.Info("Migrations ran successfully")
	return nil
}
