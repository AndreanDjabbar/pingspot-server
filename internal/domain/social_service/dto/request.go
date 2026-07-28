package dto

type FollowRequest struct {
	FollowingID   uint   `json:"followingID" validate:"required"`
	FollowingType string `json:"followingType" validate:"required,oneof=user community"`
}

type GetFollowDataRequest struct {
	FollowingID   uint   `json:"followingID" validate:"required"`
	FollowingType string `json:"followingType" validate:"required,oneof=user community"`
}