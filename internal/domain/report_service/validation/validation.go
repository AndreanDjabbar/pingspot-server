package validation

import (
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()
}

func FormatCreateReportValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "ReportTitle":
			if e.Tag() == "required" {
				errors["reportTitle"] = "Report title is required"
			}
			if e.Tag() == "max" {
				errors["reportTitle"] = "Report title must be at most 200 characters"
			}
		case "ReportType":
			if e.Tag() == "required" {
				errors["reportType"] = "Report type is required"
			}
			if e.Tag() == "oneof" {
				errors["reportType"] = "Report type must be one of INFRASTRUCTURE, ENVIRONMENT, SAFETY, OTHER"
			}
		case "ReportDescription":
			if e.Tag() == "required" {
				errors["reportDescription"] = "Report description is required"
			}
		case "DetailLocation":
			if e.Tag() == "required" {
				errors["detailLocation"] = "Location details are required"
			}
		case "HasProgress":
			if e.Tag() == "omitempty" {
				errors["hasProgress"] = "HasProgress is invalid"
			}
		case "Latitude":
			if e.Tag() == "required" {
				errors["latitude"] = "Latitude is required"
			}
		case "Longitude":
			if e.Tag() == "required" {
				errors["longitude"] = "Longitude is required"
			}
		case "DisplayName":
			if e.Tag() == "max" {
				errors["displayName"] = "Display name must be at most 255 characters"
			}
		case "AddressType":
			if e.Tag() == "max" {
				errors["addressType"] = "Address type must be at most 100 characters"
			}
		case "Country":
			if e.Tag() == "max" {
				errors["country"] = "Country must be at most 100 characters"
			}
		case "CountryCode":
			if e.Tag() == "max" {
				errors["countryCode"] = "Country code must be at most 10 characters"
			}
		case "Region":
			if e.Tag() == "max" {
				errors["region"] = "Region must be at most 100 characters"
			}
		case "PostCode":
			if e.Tag() == "max" {
				errors["postCode"] = "Postal code must be at most 20 characters"
			}
		case "County":
			if e.Tag() == "max" {
				errors["county"] = "County must be at most 200 characters"
			}
		case "State":
			if e.Tag() == "max" {
				errors["state"] = "State must be at most 200 characters"
			}
		case "Road":
			if e.Tag() == "max" {
				errors["road"] = "Road must be at most 200 characters"
			}
		case "Village":
			if e.Tag() == "max" {
				errors["village"] = "Village must be at most 200 characters"
			}
		case "Suburb":
			if e.Tag() == "max" {
				errors["suburb"] = "Suburb must be at most 200 characters"
			}
		case "Image1URL":
			if e.Tag() == "max" {
				errors["image1Url"] = "Image URL 1 must be at most 255 characters"
			}
		case "Image2URL":
			if e.Tag() == "max" {
				errors["image2Url"] = "Image URL 2 must be at most 255 characters"
			}
		case "Image3URL":
			if e.Tag() == "max" {
				errors["image3Url"] = "Image URL 3 must be at most 255 characters"
			}
		case "Image4URL":
			if e.Tag() == "max" {
				errors["image4Url"] = "Image URL 4 must be at most 255 characters"
			}
		case "Image5URL":
			if e.Tag() == "max" {
				errors["image5Url"] = "Image URL 5 must be at most 255 characters"
			}
		}
	}
	return errors
}

func FormatEditReportValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "ReportTitle":
			if e.Tag() == "required" {
				errors["reportTitle"] = "Report title is required"
			}
			if e.Tag() == "max" {
				errors["reportTitle"] = "Report title must be at most 200 characters"
			}
		case "ReportType":
			if e.Tag() == "required" {
				errors["reportType"] = "Report type is required"
			}
			if e.Tag() == "oneof" {
				errors["reportType"] = "Report type must be one of INFRASTRUCTURE, ENVIRONMENT, SAFETY, OTHER"
			}
		case "ReportDescription":
			if e.Tag() == "required" {
				errors["reportDescription"] = "Report description is required"
			}
		case "DetailLocation":
			if e.Tag() == "required" {
				errors["detailLocation"] = "Location details are required"
			}
		case "HasProgress":
			if e.Tag() == "omitempty" {
				errors["hasProgress"] = "HasProgress is invalid"
			}
		case "Latitude":
			if e.Tag() == "required" {
				errors["latitude"] = "Latitude is required"
			}
		case "Longitude":
			if e.Tag() == "required" {
				errors["longitude"] = "Longitude is required"
			}
		case "DisplayName":
			if e.Tag() == "max" {
				errors["displayName"] = "Display name must be at most 255 characters"
			}
		case "AddressType":
			if e.Tag() == "max" {
				errors["addressType"] = "Address type must be at most 100 characters"
			}
		case "Country":
			if e.Tag() == "max" {
				errors["country"] = "Country must be at most 100 characters"
			}
		case "CountryCode":
			if e.Tag() == "max" {
				errors["countryCode"] = "Country code must be at most 10 characters"
			}
		case "Region":
			if e.Tag() == "max" {
				errors["region"] = "Region must be at most 100 characters"
			}
		case "PostCode":
			if e.Tag() == "max" {
				errors["postCode"] = "Postal code must be at most 20 characters"
			}
		case "County":
			if e.Tag() == "max" {
				errors["county"] = "County must be at most 200 characters"
			}
		case "State":
			if e.Tag() == "max" {
				errors["state"] = "State must be at most 200 characters"
			}
		case "Road":
			if e.Tag() == "max" {
				errors["road"] = "Road must be at most 200 characters"
			}
		case "Village":
			if e.Tag() == "max" {
				errors["village"] = "Village must be at most 200 characters"
			}
		case "Suburb":
			if e.Tag() == "max" {
				errors["suburb"] = "Suburb must be at most 200 characters"
			}
		case "Image1URL":
			if e.Tag() == "max" {
				errors["image1Url"] = "Image URL 1 must be at most 255 characters"
			}
		case "Image2URL":
			if e.Tag() == "max" {
				errors["image2Url"] = "Image URL 2 must be at most 255 characters"
			}
		case "Image3URL":
			if e.Tag() == "max" {
				errors["image3Url"] = "Image URL 3 must be at most 255 characters"
			}
		case "Image4URL":
			if e.Tag() == "max" {
				errors["image4Url"] = "Image URL 4 must be at most 255 characters"
			}
		case "Image5URL":
			if e.Tag() == "max" {
				errors["image5Url"] = "Image URL 5 must be at most 255 characters"
			}
		}
	}
	return errors
}

func FormatReactionReportValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "ReactionType":
			if e.Tag() == "required" {
				errors["reactionType"] = "Reaction type is required"
			}
			if e.Tag() == "oneof" {
				errors["reactionType"] = "Reaction type must be one of LIKE, DISLIKE"
			}
		}
	}
	return errors
}

func FormatVoteReportValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "VoteType":
			if e.Tag() == "required" {
				errors["voteType"] = "Vote type is required"
			}
			if e.Tag() == "oneof" {
				errors["voteType"] = "Vote type must be one of RESOLVED, ON_PROGRESS, NOT_RESOLVED"
			}
		}
	}
	return errors
}

func FormatUploadProgressReportValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "Status":
			if e.Tag() == "required" {
				errors["status"] = "Status is required"
			}
			if e.Tag() == "oneof" {
				errors["status"] = "Status must be one of RESOLVED, NOT_RESOLVED, ON_PROGRESS"
			}
		case "Notes":
			if e.Tag() == "omitempty" {
				errors["notes"] = "Notes are invalid"
			}
		case "Attachment1":
			if e.Tag() == "omitempty" {
				errors["attachment1"] = "Attachment 1 is invalid"
			}
		case "Attachment2":
			if e.Tag() == "omitempty" {
				errors["attachment2"] = "Attachment 2 is invalid"
			}
		}
	}
	return errors
}

func FormatCreateReportCommentValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "Content":
			if e.Tag() == "omitempty" {
				errors["content"] = "Content is invalid"
			}
			if e.Tag() == "max" {
				errors["content"] = "Content must be at most 1000 characters"
			}
		case "MediaURL":
			if e.Tag() == "omitempty" {
				errors["mediaURL"] = "Media URL is invalid"
			}
			if e.Tag() == "max" {
				errors["mediaURL"] = "Media URL must be at most 255 characters"
			}
		case "MediaType":
			if e.Tag() == "omitempty" {
				errors["mediaType"] = "Media type is invalid"
			}
			if e.Tag() == "oneof" {
				errors["mediaType"] = "Media type must be one of IMAGE, GIF, VIDEO"
			}
		case "MediaWidth":
			if e.Tag() == "omitempty" {
				errors["mediaWidth"] = "Media width is invalid"
			}
			if e.Tag() == "min" {
				errors["mediaWidth"] = "Media width must be at least 1"
			}
		case "MediaHeight":
			if e.Tag() == "omitempty" {
				errors["mediaHeight"] = "Media height is invalid"
			}
			if e.Tag() == "min" {
				errors["mediaHeight"] = "Media height must be at least 1"
			}
		case "Mentions":
			if e.Tag() == "omitempty" {
				errors["mentions"] = "Mentions are invalid"
			}
			if e.Tag() == "dive" || e.Tag() == "gt" {
				errors["mentions"] = "Mentions must contain valid user IDs"
			}
		case "ParentCommentID":
			if e.Tag() == "omitempty" {
				errors["parentCommentID"] = "Parent comment ID is invalid"
			}
			if e.Tag() == "len" {
				errors["parentCommentID"] = "Parent comment ID must be 24 characters long"
			}
		case "ThreadRootID":
			if e.Tag() == "omitempty" {
				errors["threadRootID"] = "Thread root ID is invalid"
			}
			if e.Tag() == "len" {
				errors["threadRootID"] = "Thread root ID must be 24 characters long"
			}
		}
	}
	return errors
}

func FormatSaveReportValidationErrors(err error) map[string]string {
	errors := map[string]string{}
	if err == nil {
		return errors
	}
	for _, e := range err.(validator.ValidationErrors) {
		switch e.Field() {
		case "Save":
			if e.Tag() == "required" {
				errors["save"] = "Save field is required"
			}
		}
	}
	return errors
}
