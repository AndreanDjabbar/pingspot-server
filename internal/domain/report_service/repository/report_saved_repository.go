package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type ReportSavedRepository interface {
	CreateTX(ctx context.Context, tx *gorm.DB, reportSaved *model.ReportSaved) error
	GetByUserIDAndReportID(ctx context.Context, userID uint, reportID uint) (*model.ReportSaved, error)
	DeleteTX(ctx context.Context, tx *gorm.DB, reportSaved *model.ReportSaved) error
}

type reportSavedRepository struct {
	db *gorm.DB
}

func NewReportSavedRepository(db *gorm.DB) ReportSavedRepository {
	return &reportSavedRepository{db: db}
}

func (r *reportSavedRepository) CreateTX(ctx context.Context, tx *gorm.DB, reportSaved *model.ReportSaved) error {
	if err := tx.WithContext(ctx).Create(reportSaved).Error; err != nil {
		return err
	}
	return nil
}

func (r *reportSavedRepository) GetByUserIDAndReportID(ctx context.Context, userID uint, reportID uint) (*model.ReportSaved, error) {
	var reportSaved model.ReportSaved
	if err := r.db.WithContext(ctx).Where("user_id = ? AND report_id = ?", userID, reportID).First(&reportSaved).Error; err != nil {
		return nil, err
	}
	return &reportSaved, nil
}

func (r *reportSavedRepository) DeleteTX(ctx context.Context, tx *gorm.DB, reportSaved *model.ReportSaved) error {
	if err := tx.WithContext(ctx).Delete(reportSaved).Error; err != nil {
		return err
	}
	return nil
}