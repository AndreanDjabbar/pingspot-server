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
					errors["reportTitle"] = "Judul laporan wajib diisi"
				}
				if e.Tag() == "max" {
					errors["reportTitle"] = "Judul laporan maksimal 200 karakter"
				}
			case "ReportType":
				if e.Tag() == "required" {
					errors["reportType"] = "Tipe laporan wajib diisi"
				}
				if e.Tag() == "oneof" {
					errors["reportType"] = "Tipe laporan harus salah satu antara INFRASTRUCTURE, ENVIRONMENT, SAFETY, OTHER"
				}
			case "ReportDescription":
				if e.Tag() == "required" {
					errors["reportDescription"] = "Deskripsi laporan wajib diisi"
				}
			case "DetailLocation":
				if e.Tag() == "required" {
					errors["detailLocation"] = "Detail lokasi wajib diisi"
				}
			case "HasProgress":
				if e.Tag() == "omitempty" {
					errors["hasProgress"] = "HasProgress tidak valid"
				}
			case "Latitude":
				if e.Tag() == "required" {
					errors["latitude"] = "Latitude wajib diisi"
				}
			case "Longitude":
				if e.Tag() == "required" {
					errors["longitude"] = "Longitude wajib diisi"
				}
			case "DisplayName":
				if e.Tag() == "max" {
					errors["displayName"] = "Display name maksimal 255 karakter"
				}
			case "AddressType":
				if e.Tag() == "max" {
					errors["addressType"] = "Tipe alamat maksimal 100 karakter"
				}
			case "Country":
				if e.Tag() == "max" {
					errors["country"] = "Negara maksimal 100 karakter"
				}
			case "CountryCode":
				if e.Tag() == "max" {
					errors["countryCode"] = "Kode negara maksimal 10 karakter"
				}
			case "Region":
				if e.Tag() == "max" {
					errors["region"] = "Region maksimal 100 karakter"
				}
			case "PostCode":
				if e.Tag() == "max" {
					errors["postCode"] = "Kode pos maksimal 20 karakter"
				}
			case "County":
				if e.Tag() == "max" {
					errors["county"] = "County maksimal 200 karakter"
				}
			case "State":
				if e.Tag() == "max" {
					errors["state"] = "State maksimal 200 karakter"
				}
			case "Road":
				if e.Tag() == "max" {
					errors["road"] = "Road maksimal 200 karakter"
				}
			case "Village":
				if e.Tag() == "max" {
					errors["village"] = "Village maksimal 200 karakter"
				}
			case "Suburb":
				if e.Tag() == "max" {
					errors["suburb"] = "Suburb maksimal 200 karakter"
				}
			case "Image1URL":
				if e.Tag() == "max" {
					errors["image1Url"] = "URL gambar 1 maksimal 255 karakter"
				}
			case "Image2URL":
				if e.Tag() == "max" {
					errors["image2Url"] = "URL gambar 2 maksimal 255 karakter"
				}
			case "Image3URL":
				if e.Tag() == "max" {
					errors["image3Url"] = "URL gambar 3 maksimal 255 karakter"
				}
			case "Image4URL":
				if e.Tag() == "max" {
					errors["image4Url"] = "URL gambar 4 maksimal 255 karakter"
				}
			case "Image5URL":
				if e.Tag() == "max" {
					errors["image5Url"] = "URL gambar 5 maksimal 255 karakter"
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
					errors["reportTitle"] = "Judul laporan wajib diisi"
				}
				if e.Tag() == "max" {
					errors["reportTitle"] = "Judul laporan maksimal 200 karakter"
				}
			case "ReportType":
				if e.Tag() == "required" {
					errors["reportType"] = "Tipe laporan wajib diisi"
				}
				if e.Tag() == "oneof" {
					errors["reportType"] = "Tipe laporan harus salah satu antara INFRASTRUCTURE, ENVIRONMENT, SAFETY, OTHER"
				}
			case "ReportDescription":
				if e.Tag() == "required" {
					errors["reportDescription"] = "Deskripsi laporan wajib diisi"
				}
			case "DetailLocation":
				if e.Tag() == "required" {
					errors["detailLocation"] = "Detail lokasi wajib diisi"
				}
			case "HasProgress":
				if e.Tag() == "omitempty" {
					errors["hasProgress"] = "HasProgress tidak valid"
				}
			case "Latitude":
				if e.Tag() == "required" {
					errors["latitude"] = "Latitude wajib diisi"
				}
			case "Longitude":
				if e.Tag() == "required" {
					errors["longitude"] = "Longitude wajib diisi"
				}
			case "DisplayName":
				if e.Tag() == "max" {
					errors["displayName"] = "Display name maksimal 255 karakter"
				}
			case "AddressType":
				if e.Tag() == "max" {
					errors["addressType"] = "Tipe alamat maksimal 100 karakter"
				}
			case "Country":
				if e.Tag() == "max" {
					errors["country"] = "Negara maksimal 100 karakter"
				}
			case "CountryCode":
				if e.Tag() == "max" {
					errors["countryCode"] = "Kode negara maksimal 10 karakter"
				}
			case "Region":
				if e.Tag() == "max" {
					errors["region"] = "Region maksimal 100 karakter"
				}
			case "PostCode":
				if e.Tag() == "max" {
					errors["postCode"] = "Kode pos maksimal 20 karakter"
				}
			case "County":
				if e.Tag() == "max" {
					errors["county"] = "County maksimal 200 karakter"
				}
			case "State":
				if e.Tag() == "max" {
					errors["state"] = "State maksimal 200 karakter"
				}
			case "Road":
				if e.Tag() == "max" {
					errors["road"] = "Road maksimal 200 karakter"
				}
			case "Village":
				if e.Tag() == "max" {
					errors["village"] = "Village maksimal 200 karakter"
				}
			case "Suburb":
				if e.Tag() == "max" {
					errors["suburb"] = "Suburb maksimal 200 karakter"
				}
			case "Image1URL":
				if e.Tag() == "max" {
					errors["image1Url"] = "URL gambar 1 maksimal 255 karakter"
				}
			case "Image2URL":
				if e.Tag() == "max" {
					errors["image2Url"] = "URL gambar 2 maksimal 255 karakter"
				}
			case "Image3URL":
				if e.Tag() == "max" {
					errors["image3Url"] = "URL gambar 3 maksimal 255 karakter"
				}
			case "Image4URL":
				if e.Tag() == "max" {
					errors["image4Url"] = "URL gambar 4 maksimal 255 karakter"
				}
			case "Image5URL":
				if e.Tag() == "max" {
					errors["image5Url"] = "URL gambar 5 maksimal 255 karakter"
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
					errors["reactionType"] = "Tipe reaksi wajib diisi"
				}
				if e.Tag() == "oneof" {
					errors["reactionType"] = "Tipe reaksi harus salah satu antara LIKE, DISLIKE"
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
					errors["voteType"] = "Tipe vote wajib diisi"
				}
				if e.Tag() == "oneof" {
					errors["voteType"] = "Tipe vote harus salah satu antara RESOLVED, ON_PROGRESS, NOT_RESOLVED"
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
					errors["status"] = "Status wajib diisi"
				}
				if e.Tag() == "oneof" {
					errors["status"] = "Status harus salah satu antara RESOLVED, NOT_RESOLVED, ON_PROGRESS"
				}
			case "Notes":
				if e.Tag() == "omitempty" {
					errors["notes"] = "Catatan tidak valid"
				}	
			case "Attachment1":
				if e.Tag() == "omitempty" {
					errors["attachment1"] = "Attachment 1 tidak valid"
				}
			case "Attachment2":
				if e.Tag() == "omitempty" {
					errors["attachment2"] = "Attachment 2 tidak valid"
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
				errors["content"] = "Konten tidak valid"
			}
			if e.Tag() == "max" {
				errors["content"] = "Konten maksimal 1000 karakter"
			}
		case "MediaURL":
			if e.Tag() == "omitempty" {
				errors["mediaURL"] = "Media URL tidak valid"
			}
			if e.Tag() == "max" {
				errors["mediaURL"] = "Media URL maksimal 255 karakter"
			}
		case "MediaType":
			if e.Tag() == "omitempty" {
				errors["mediaType"] = "Tipe media tidak valid"
			}
			if e.Tag() == "oneof" {
				errors["mediaType"] = "Tipe media harus salah satu antara IMAGE, GIF, VIDEO"
			}
		case "MediaWidth":
			if e.Tag() == "omitempty" {
				errors["mediaWidth"] = "Media width tidak valid"
			}
			if e.Tag() == "min" {
				errors["mediaWidth"] = "Media width minimal 1"
			}
		case "MediaHeight":
			if e.Tag() == "omitempty" {
				errors["mediaHeight"] = "Media height tidak valid"
			}
			if e.Tag() == "min" {
				errors["mediaHeight"] = "Media height minimal 1"
			}
		case "Mentions":
			if e.Tag() == "omitempty" {
				errors["mentions"] = "Mentions tidak valid"
			}
			if e.Tag() == "dive" || e.Tag() == "gt" {
				errors["mentions"] = "Mentions harus berisi ID user yang valid"
			}
		case "ParentCommentID":
			if e.Tag() == "omitempty" {
				errors["parentCommentID"] = "Parent comment ID tidak valid"
			}
			if e.Tag() == "len" {
				errors["parentCommentID"] = "Parent comment ID harus memiliki panjang 24 karakter"
			}
		case "ThreadRootID":
			if e.Tag() == "omitempty" {
				errors["threadRootID"] = "Thread root ID tidak valid"
			}
			if e.Tag() == "len" {
				errors["threadRootID"] = "Thread root ID harus memiliki panjang 24 karakter"
			}
		}
	}
	return errors
}