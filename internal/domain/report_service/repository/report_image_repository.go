package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type ReportImageRepository interface {
	Create(ctx context.Context, images *model.ReportImage, tx *gorm.DB) error
	UpdateTX(ctx context.Context, tx *gorm.DB, images *model.ReportImage) (*model.ReportImage, error)
	GetByReportID(ctx context.Context, reportID uint) (*model.ReportImage, error)
}

type reportImageRepository struct {
	db *gorm.DB
}

func NewReportImageRepository(db *gorm.DB) ReportImageRepository {
	return &reportImageRepository{db: db}
}

func (r *reportImageRepository) Create(ctx context.Context, images *model.ReportImage, tx *gorm.DB) error {
	if tx != nil {
		return tx.WithContext(ctx).Create(images).Error
	}
	return r.db.WithContext(ctx).Create(images).Error
}

func (r *reportImageRepository) UpdateTX(ctx context.Context, tx *gorm.DB, images *model.ReportImage) (*model.ReportImage, error) {
	if err := tx.WithContext(ctx).Save(images).Error; err != nil {
		return nil, err
	}
	return images, nil
}

func (r *reportImageRepository) GetByReportID(ctx context.Context, reportID uint) (*model.ReportImage, error) {
	var images model.ReportImage
	if err := r.db.WithContext(ctx).Where("report_id = ?", reportID).First(&images).Error; err != nil {
		return nil, err
	}
	return &images, nil
}
