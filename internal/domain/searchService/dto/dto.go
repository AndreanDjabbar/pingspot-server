package dto

type UsersSearch struct {
	UserID          uint    `json:"userID"`
	FullName        string  `json:"fullName"`
	Bio             *string `json:"bio"`
	ProfilePicture  *string `json:"profilePicture"`
	Username		string  `json:"username"`
	Birthday   		*string `json:"birthday"`
	Gender 	   		*string `json:"gender"`
	Email			string  `json:"email"`	
}

type ReportsSearch struct {
	ID		  uint    `json:"id"`
	ReportTitle   string  `json:"reportTitle"`
	ReportType    string  `json:"reportType"`
	ReportDescription   string  `json:"reportDescription"`
	ReportHasProgress	bool	`json:"hasProgress"`
	ReportStatus  string  `json:"reportStatus"`
	CreatedAt     int64  `json:"reportCreatedAt"`
	UpdatedAt     int64  `json:"reportUpdatedAt"`
}

type UserSearchResult struct {
	Users []UsersSearch `json:"users"`
	Type  string       `json:"type"`
}

type ReportSearchResult struct {
	Reports []ReportsSearch `json:"reports"`
	Type    string         `json:"type"`
}
