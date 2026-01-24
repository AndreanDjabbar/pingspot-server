package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type ReportProgressRepository interface {
	Create(ctx context.Context, progress *model.ReportProgress) (*model.ReportProgress, error)
	CreateTX(ctx context.Context, tx *gorm.DB, progress *model.ReportProgress) (*model.ReportProgress, error)
	GetByReportID(ctx context.Context, reportID uint) ([]model.ReportProgress, error)
}

type reporProgressRepository struct {
	db *gorm.DB
}

func NewReportProgressRepository(db *gorm.DB) ReportProgressRepository {
	return &reporProgressRepository{db: db}
}

func (r *reporProgressRepository) Create(ctx context.Context, progress *model.ReportProgress) (*model.ReportProgress, error) {
	if err := r.db.WithContext(ctx).Create(progress).Error; err != nil {
		return nil, err
	}
	return progress, nil
}

func (r *reporProgressRepository) CreateTX(ctx context.Context, tx *gorm.DB, progress *model.ReportProgress) (*model.ReportProgress, error) {
	if err := tx.WithContext(ctx).Create(progress).Error; err != nil {
		return nil, err
	}
	return progress, nil
}

func (r *reporProgressRepository) GetByReportID(ctx context.Context, reportID uint) ([]model.ReportProgress, error) {
	var progresses []model.ReportProgress
	if err := r.db.WithContext(ctx).Where("report_id = ?", reportID).Preload("User").Find(&progresses).Error; err != nil {
		return nil, err
	}
	return progresses, nil
}
