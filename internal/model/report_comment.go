package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type CommentMediaType string

const (
	Image CommentMediaType = "IMAGE"
	Gif   CommentMediaType = "GIF"
	Video CommentMediaType = "VIDEO"
)

type CommentMedia struct {
	URL 	string `bson:"url" json:"url"`
	Type	CommentMediaType `bson:"type" json:"type"`
	Width 	*uint    `bson:"width,omitempty" json:"width,omitempty"`
	Height 	*uint    `bson:"height,omitempty" json:"height,omitempty"`
}

type Mention struct {
	UserID   uint   `bson:"user_id" json:"userID"`
	Username string `bson:"username" json:"username"`
}

type ReportComment struct {
	ID primitive.ObjectID `bson:"_id,omitempty"`
	ReportID  uint 		`bson:"report_id"`
	UserID    uint 		`bson:"user_id"`

	Content  *string    `bson:"content,omitempty"`
	Media *CommentMedia `bson:"media,omitempty"`
	Mentions []Mention    `bson:"mentions,omitempty"`

	ParentCommentID *primitive.ObjectID `bson:"parent_comment_id,omitempty"`
	ThreadRootID    *primitive.ObjectID `bson:"thread_root_id,omitempty"`

	CreatedAt int64 `bson:"created_at"`
	UpdatedAt *int64 `bson:"updated_at,omitempty"`
}