package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type ReportReactionRepository interface {
	WithTX(tx *gorm.DB) ReportReactionRepository
	GetByUserReportID(ctx context.Context, userID, reportID uint) (*model.ReportReaction, error)
	Create(ctx context.Context, reaction *model.ReportReaction) (*model.ReportReaction, error)
	Update(ctx context.Context, reaction *model.ReportReaction) (*model.ReportReaction, error)
	Delete(ctx context.Context, reaction *model.ReportReaction) error
	GetLikeReactionCount(ctx context.Context, reportID uint) (int64, error)
	GetDislikeReactionCount(ctx context.Context, reportID uint) (int64, error)
}

type reportReactionRepository struct {
	db *gorm.DB
}

func NewReportReactionRepository(db *gorm.DB) ReportReactionRepository {
	return &reportReactionRepository{db: db}
}

func (r *reportReactionRepository) WithTX(tx *gorm.DB) ReportReactionRepository {
	return &reportReactionRepository{
		db: tx,
	}
}

func (r *reportReactionRepository) GetByUserReportID(ctx context.Context, userID, reportID uint) (*model.ReportReaction, error) {
	var reaction model.ReportReaction
	if err := r.db.WithContext(ctx).Where("user_id = ? AND report_id = ?", userID, reportID).First(&reaction).Error; err != nil {
		return nil, err
	}
	return &reaction, nil
}

func (r *reportReactionRepository) Create(ctx context.Context, reaction *model.ReportReaction) (*model.ReportReaction, error) {
	if err := r.db.WithContext(ctx).Create(reaction).Error; err != nil {
		return nil, err
	}
	return reaction, nil
}

func (r *reportReactionRepository) Update(ctx context.Context, reaction *model.ReportReaction) (*model.ReportReaction, error) {
	if err := r.db.WithContext(ctx).Save(reaction).Error; err != nil {
		return nil, err
	}
	return reaction, nil
}

func (r *reportReactionRepository) Delete(ctx context.Context, reaction *model.ReportReaction) error {
	if err := r.db.WithContext(ctx).Delete(reaction).Error; err != nil {
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
