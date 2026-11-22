package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type ReportLocationRepository interface {
	WithTX(tx *gorm.DB) ReportLocationRepository
	Create(ctx context.Context, location *model.ReportLocation, tx *gorm.DB) error
	Update(ctx context.Context, location *model.ReportLocation) (*model.ReportLocation, error)
	GetByReportID(ctx context.Context, reportID uint) (*model.ReportLocation, error)
}

type reportLocationRepository struct {
	db *gorm.DB
}

func NewReportLocationRepository(db *gorm.DB) ReportLocationRepository {
	return &reportLocationRepository{db: db}
}

func (r *reportLocationRepository) WithTX(tx *gorm.DB) ReportLocationRepository {
	return &reportLocationRepository{
		db: tx,
	}
}

func (r *reportLocationRepository) Create(ctx context.Context, location *model.ReportLocation, tx *gorm.DB) error {
	if tx != nil {
		return tx.WithContext(ctx).Create(location).Error
	}
	return r.db.WithContext(ctx).Create(location).Error
}

func (r *reportLocationRepository) Update(ctx context.Context, location *model.ReportLocation) (*model.ReportLocation, error) {
	if err := r.db.WithContext(ctx).Save(location).Error; err != nil {
		return nil, err
	}
	return location, nil
}

func (r *reportLocationRepository) GetByReportID(ctx context.Context, reportID uint) (*model.ReportLocation, error) {
	var location model.ReportLocation
	if err := r.db.WithContext(ctx).Where("report_id = ?", reportID).First(&location).Error; err != nil {
		return nil, err
	}
	return &location, nil
}
