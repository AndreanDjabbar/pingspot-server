package repository

import (
	"context"
	"pingspot/internal/domain/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ReportCommentRepository interface {
	Create(ctx context.Context, report *model.ReportComment) (*model.ReportComment, error)
	GetByID(ctx context.Context, commentID primitive.ObjectID) (*model.ReportComment, error)
	GetByIDs(ctx context.Context, commentIDs []primitive.ObjectID) ([]*model.ReportComment, error)
	GetByReportID(ctx context.Context, reportID uint) ([]*model.ReportComment, error)
	GetCountsByReportID(ctx context.Context, reportID uint) (int64, error)
	GetCountsByRootID(ctx context.Context, rootID primitive.ObjectID) (int64, error)
	GetPaginatedRootByReportID(ctx context.Context, reportID uint, cursorID *primitive.ObjectID, limit int) ([]*model.ReportComment, error)
	GetPaginatedRepliesByRootID(ctx context.Context, rootID primitive.ObjectID, cursorID *primitive.ObjectID, limit int) ([]*model.ReportComment, error)
}

type reportCommentRepository struct {
	db         *mongo.Client
	collection *mongo.Collection
}

func NewReportCommentRepository(db *mongo.Client) ReportCommentRepository {
	return &reportCommentRepository{
		db:         db,
		collection: db.Database("report_service").Collection("report_comments"),
	}
}

func (r *reportCommentRepository) Create(ctx context.Context, comment *model.ReportComment) (*model.ReportComment, error) {
	comment.ID = primitive.NewObjectID()

	_, err := r.collection.InsertOne(ctx, comment)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (r *reportCommentRepository) GetByID(ctx context.Context, commentID primitive.ObjectID) (*model.ReportComment, error) {
	filter := bson.M{
		"_id": commentID,
	}
	var comment model.ReportComment
	err := r.collection.FindOne(ctx, filter).Decode(&comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *reportCommentRepository) GetByIDs(ctx context.Context, commentIDs []primitive.ObjectID) ([]*model.ReportComment, error) {
	filter := bson.M{
		"_id": bson.M{
			"$in": commentIDs,
		},
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var comments []*model.ReportComment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *reportCommentRepository) GetCountsByReportID(ctx context.Context, reportID uint) (int64, error) {
	filter := bson.M{
		"report_id": reportID,
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *reportCommentRepository) GetCountsByRootID(ctx context.Context, rootID primitive.ObjectID) (int64, error) {
	filter := bson.M{
		"thread_root_id": rootID,
	}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *reportCommentRepository) GetByReportID(ctx context.Context, reportID uint) ([]*model.ReportComment, error) {
	filter := bson.M{
		"report_id": reportID,
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var comments []*model.ReportComment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *reportCommentRepository) GetPaginatedRootByReportID(ctx context.Context, reportID uint, cursorID *primitive.ObjectID, limit int) ([]*model.ReportComment, error) {
	filter := bson.M{
		"report_id":         reportID,
		"parent_comment_id": bson.M{"$exists": false},
	}

	if cursorID != nil {
		filter["_id"] = bson.M{
			"$gt": *cursorID,
		}
	}

	findOpts := options.Find()
	if limit > 0 {
		findOpts.SetLimit(int64(limit))
	}
	findOpts.SetSort(bson.M{"_id": 1})

	cursor, err := r.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []*model.ReportComment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *reportCommentRepository) GetPaginatedRepliesByRootID(ctx context.Context, rootID primitive.ObjectID, cursorID *primitive.ObjectID, limit int) ([]*model.ReportComment, error) {
	filter := bson.M{
		"thread_root_id": rootID,
	}
	if cursorID != nil {
		filter["_id"] = bson.M{
			"$gt": *cursorID,
		}
	}
	findOpts := options.Find()
	if limit > 0 {
		findOpts.SetLimit(int64(limit))
	}
	findOpts.SetSort(bson.M{"_id": 1})

	cursor, err := r.collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var comments []*model.ReportComment
	if err := cursor.All(ctx, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}
