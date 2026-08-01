package dto

type FollowResponse struct {
	FollowingID    uint   `json:"followingID"`
	FollowingType  string `json:"followingType"`
	FollowerUserID uint   `json:"followerUserID"`
	FollowProcess  string `json:"followProcess"`
}

type GetFollowDataResponse struct {
	FollowingID    uint          `json:"followingID"`
	FollowersCount int64         `json:"followersCount"`
	FollowingCount int64         `json:"followingCount"`
	MyFollowData   *Follow `json:"myFollowData,omitempty"`
}