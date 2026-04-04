package dto

type SaveUserProfileResponse struct {
	UserID          uint    `json:"userID"`
	FullName        string  `json:"fullName"`
	Bio             *string `json:"bio"`
	ProfilePicture  *string `json:"profilePicture"`
	Username		string  `json:"username"`
	Gender 	   		*string `json:"gender"`
	Birthday   		*string `json:"birthday"`
}

type GetProfileResponse struct {
	UserID          uint    `json:"userID"`
	FullName        string  `json:"fullName"`
	Bio             *string `json:"bio"`
	ProfilePicture  *string `json:"profilePicture"`
	Username		string  `json:"username"`
	Birthday   		*string `json:"birthday"`
	Gender 	   		*string `json:"gender"`
	Email			string  `json:"email"`	
}

type GetUserStatisticsResponse struct {
	TotalUsers         int64            `json:"totalUsers"`
	UsersByGender      map[string]int64 `json:"usersByGender"`
	MonthlyUserCounts  map[string]int64 `json:"monthlyUserCounts"`
}