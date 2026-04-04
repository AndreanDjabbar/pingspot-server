package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type ReportReactionRepository interface {
	GetByUserReportID(ctx context.Context, userID, reportID uint) (*model.ReportReaction, error)
	GetByUserReportIDTX(ctx context.Context, tx *gorm.DB, userID, reportID uint) (*model.ReportReaction, error)
	CreateTX(ctx context.Context, tx *gorm.DB, reaction *model.ReportReaction) (*model.ReportReaction, error)
	UpdateTX(ctx context.Context, tx *gorm.DB, reaction *model.ReportReaction) (*model.ReportReaction, error)
	DeleteTX(ctx context.Context, tx *gorm.DB, reaction *model.ReportReaction) error
	GetLikeReactionCount(ctx context.Context, reportID uint) (int64, error)
	GetDislikeReactionCount(ctx context.Context, reportID uint) (int64, error)
}

type reportReactionRepository struct {
	db *gorm.DB
}

func NewReportReactionRepository(db *gorm.DB) ReportReactionRepository {
	return &reportReactionRepository{db: db}
}

func (r *reportReactionRepository) GetByUserReportID(ctx context.Context, userID, reportID uint) (*model.ReportReaction, error) {
	var reaction model.ReportReaction
	if err := r.db.WithContext(ctx).Where("user_id = ? AND report_id = ?", userID, reportID).First(&reaction).Error; err != nil {
		return nil, err
	}
	return &reaction, nil
}

func (r *reportReactionRepository) GetByUserReportIDTX(ctx context.Context, tx *gorm.DB, userID, reportID uint) (*model.ReportReaction, error) {
	var reaction model.ReportReaction
	if err := tx.WithContext(ctx).Where("user_id = ? AND report_id = ?", userID, reportID).First(&reaction).Error; err != nil {
		return nil, err
	}
	return &reaction, nil
}

func (r *reportReactionRepository) CreateTX(ctx context.Context, tx *gorm.DB, reaction *model.ReportReaction) (*model.ReportReaction, error) {
	if err := tx.WithContext(ctx).Create(reaction).Error; err != nil {
		return nil, err
	}
	return reaction, nil
}

func (r *reportReactionRepository) UpdateTX(ctx context.Context, tx *gorm.DB, reaction *model.ReportReaction) (*model.ReportReaction, error) {
	if err := tx.WithContext(ctx).Save(reaction).Error; err != nil {
		return nil, err
	}
	return reaction, nil
}

func (r *reportReactionRepository) DeleteTX(ctx context.Context, tx *gorm.DB, reaction *model.ReportReaction) error {
	if err := tx.WithContext(ctx).Delete(reaction).Error; err != nil {
		return err
	}
	return nil
}

func (r *reportReactionRepository) GetLikeReactionCount(ctx context.Context, reportID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ReportReaction{}).
		Where("report_id = ? AND type = ?", reportID, model.Like).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *reportReactionRepository) GetDislikeReactionCount(ctx context.Context, reportID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ReportReaction{}).
		Where("report_id = ? AND type = ?", reportID, model.Dislike).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
