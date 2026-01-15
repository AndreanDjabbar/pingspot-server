package dto

type SearchResponse struct {
	UsersData UserSearchResult	`json:"usersData"`
	ReportsData ReportSearchResult `json:"reportsData"`
}