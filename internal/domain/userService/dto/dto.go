package dto

type UserProfile struct {
	UserID          uint    `json:"userID"`
	FullName        string  `json:"fullName"`
	Bio             *string `json:"bio"`
	ProfilePicture  *string `json:"profilePicture"`
	Username		string  `json:"username"`
	Gender 	   		*string `json:"gender"`
	Birthday   		*string `json:"birthday"`
}