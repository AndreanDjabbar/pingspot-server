package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type ReportVoteRepository interface {
	GetByUserReportID(ctx context.Context, userID, reportID uint) (*model.ReportVote, error)
	GetByUserReportIDTX(ctx context.Context, tx *gorm.DB, userID, reportID uint) (*model.ReportVote, error)
	CreateTX(ctx context.Context, tx *gorm.DB, vote *model.ReportVote) (*model.ReportVote, error)
	UpdateTX(ctx context.Context, tx *gorm.DB, vote *model.ReportVote) (*model.ReportVote, error)
	DeleteTX(ctx context.Context, tx *gorm.DB, vote *model.ReportVote) error
	GetReportVoteCount(ctx context.Context, voteType model.ReportStatus, reportID uint) (int64, error)
	GetReportVoteCountsTX(ctx context.Context, tx *gorm.DB, reportID uint) (map[model.ReportStatus]int64, error)
	GetHighestVoteTypeTX(ctx context.Context, tx *gorm.DB, reportID uint) (model.ReportStatus, error)
	GetResolvedVoteCount(ctx context.Context, reportID uint) (int64, error)
	GetOnProgressVoteCount(ctx context.Context, reportID uint) (int64, error)
	GetNotResolvedVoteCount(ctx context.Context, reportID uint) (int64, error)
	GetTotalVoteCountTX(ctx context.Context, tx *gorm.DB, reportID uint) (int64, error)
}

type reportVoteRepository struct {
	db *gorm.DB
}

func NewReportVoteRepository(db *gorm.DB) ReportVoteRepository {
	return &reportVoteRepository{db: db}
}

func (r *reportVoteRepository) GetByUserReportID(ctx context.Context, userID, reportID uint) (*model.ReportVote, error) {
	var vote model.ReportVote
	if err := r.db.WithContext(ctx).Where("user_id = ? AND report_id = ?", userID, reportID).First(&vote).Error; err != nil {
		return nil, err
	}
	return &vote, nil
}

func (r *reportVoteRepository) GetTotalVoteCountTX(ctx context.Context, tx *gorm.DB, reportID uint) (int64, error) {
	var count int64
	if err := tx.WithContext(ctx).Model(&model.ReportVote{}).
		Where("report_id = ?", reportID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *reportVoteRepository) GetReportVoteCountsTX(ctx context.Context, tx *gorm.DB, reportID uint) (map[model.ReportStatus]int64, error) {
	rows, err := tx.WithContext(ctx).Model(&model.ReportVote{}).
		Select("vote_type, COUNT(*) as count").
		Where("report_id = ?", reportID).
		Group("vote_type").
		Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[model.ReportStatus]int64{
		model.RESOLVED:     0,
		model.ON_PROGRESS:  0,
		model.NOT_RESOLVED: 0,
	}
	for rows.Next() {
		var voteType model.ReportStatus
		var count int64
		if err := rows.Scan(&voteType, &count); err != nil {
			return nil, err
		}
		counts[voteType] = count
	}
	return counts, nil
}

func (r *reportVoteRepository) GetByUserReportIDTX(ctx context.Context, tx *gorm.DB, userID, reportID uint) (*model.ReportVote, error) {
	var vote model.ReportVote
	if err := tx.WithContext(ctx).Where("user_id = ? AND report_id = ?", userID, reportID).First(&vote).Error; err != nil {
		return nil, err
	}
	return &vote, nil
}

func (r *reportVoteRepository) GetHighestVoteTypeTX(ctx context.Context, tx *gorm.DB, reportID uint) (model.ReportStatus, error) {
	var result struct {
		VoteType model.ReportStatus
		Count    int64
	}
	if err := tx.WithContext(ctx).Model(&model.ReportVote{}).
		Select("vote_type, COUNT(*) as count").
		Where("report_id = ?", reportID).
		Group("vote_type").
		Order("count DESC").
		Limit(1).
		Scan(&result).Error; err != nil {
		return "", err
	}
	return result.VoteType, nil
}

func (r *reportVoteRepository) CreateTX(ctx context.Context, tx *gorm.DB, vote *model.ReportVote) (*model.ReportVote, error) {
	if err := tx.WithContext(ctx).Create(vote).Error; err != nil {
		return nil, err
	}
	return vote, nil
}

func (r *reportVoteRepository) UpdateTX(ctx context.Context, tx *gorm.DB, vote *model.ReportVote) (*model.ReportVote, error) {
	if err := tx.WithContext(ctx).Save(vote).Error; err != nil {
		return nil, err
	}
	return vote, nil
}

func (r *reportVoteRepository) DeleteTX(ctx context.Context, tx *gorm.DB, vote *model.ReportVote) error {
	if err := tx.WithContext(ctx).Delete(vote).Error; err != nil {
		return err
	}
	return nil
}

func (r *reportVoteRepository) GetResolvedVoteCount(ctx context.Context, reportID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ReportVote{}).
		Where("report_id = ? AND vote_type = ?", reportID, model.RESOLVED).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *reportVoteRepository) GetReportVoteCount(ctx context.Context, voteType model.ReportStatus, reportID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ReportVote{}).
		Where("report_id = ? AND vote_type = ?", reportID, voteType).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *reportVoteRepository) GetOnProgressVoteCount(ctx context.Context, reportID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ReportVote{}).
		Where("report_id = ? AND vote_type = ?", reportID, model.ON_PROGRESS).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *reportVoteRepository) GetNotResolvedVoteCount(ctx context.Context, reportID uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ReportVote{}).
		Where("report_id = ? AND vote_type = ?", reportID, model.NOT_RESOLVED).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
