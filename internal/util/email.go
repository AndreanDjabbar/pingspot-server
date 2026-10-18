package util

import (
	"bytes"
	"fmt"
	"html/template"

	"gopkg.in/gomail.v2"
)

type EmailType string

const (
	EmailTypeVerification     EmailType = "verification"
	EmailTypePasswordReset    EmailType = "password_reset"
	EmailTypeProgressReminder EmailType = "progress_reminder"
	EmailTypeNewReport        EmailType = "new_report"
	EmailTypeReportComment     EmailType = "report_comment"
)


type EmailData struct {
	To            string
	Subject       string
	RecipientName string
	EmailType     EmailType
	DisabledNotifications bool
	BodyTempate   string
	TemplateData  map[string]any
}

func RenderEmailTemplate(data EmailData) (string, error) {
	var templateHTML string
	var templateData any

	switch data.EmailType {
	case EmailTypeVerification:
		verificationLink, ok := data.TemplateData["VerificationLink"].(string)
		if !ok || verificationLink == "" {
			return "", fmt.Errorf("verification link is required for verification email")
		}
		templateHTML = data.BodyTempate
		templateData = struct {
			UserName         string
			VerificationLink string
		}{
			UserName:         data.RecipientName,
			VerificationLink: verificationLink,
		}

	case EmailTypePasswordReset:
		resetLink, ok := data.TemplateData["ResetLink"].(string)
		if !ok || resetLink == "" {
			return "", fmt.Errorf("reset link is required for password reset email")
		}
		templateHTML = data.BodyTempate
		templateData = struct {
			UserName  string
			ResetLink string
		}{
			UserName:  data.RecipientName,
			ResetLink: resetLink,
		}

	case EmailTypeProgressReminder:
		reportTitle, ok := data.TemplateData["ReportTitle"].(string)
		if !ok || reportTitle == "" {
			return "", fmt.Errorf("report title is required for progress reminder email")
		}
		reportLink, ok := data.TemplateData["ReportLink"].(string)
		if !ok || reportLink == "" {
			return "", fmt.Errorf("report link is required for progress reminder email")
		}
		daysRemaining, ok := data.TemplateData["DaysRemaining"].(int)
		if !ok {
			return "", fmt.Errorf("days remaining is required for progress reminder email")
		}
		templateHTML = data.BodyTempate
		templateData = struct {
			UserName      string
			ReportTitle   string
			ReportLink    string
			DaysRemaining int
		}{
			UserName:      data.RecipientName,
			ReportTitle:   reportTitle,
			ReportLink:    reportLink,
			DaysRemaining: daysRemaining,
		}
	
	case EmailTypeNewReport:
		reportTitle, ok := data.TemplateData["ReportTitle"].(string)
		if !ok || reportTitle == "" {
			return "", fmt.Errorf("report title is required for new report email")
		}
		reportLink, ok := data.TemplateData["ReportLink"].(string)
		if !ok || reportLink == "" {
			return "", fmt.Errorf("report link is required for new report email")
		}
		userEmail, ok := data.TemplateData["UserEmail"].(string)
		if !ok || userEmail == "" {
			return "", fmt.Errorf("user email is required for new report email")
		}

		templateHTML = data.BodyTempate
		templateData = struct {
			UserName    string
			ReportTitle string
			UserEmail   string
			ReportLink  string
		}{
			UserName:    data.RecipientName,
			ReportTitle: reportTitle,
			ReportLink:  reportLink,
			UserEmail:   userEmail,
		}

	case EmailTypeReportComment:
		reportTitle, ok := data.TemplateData["ReportTitle"].(string)
		if !ok || reportTitle == "" {
			return "", fmt.Errorf("report title is required for report comment email")
		}
		reportLink, ok := data.TemplateData["ReportLink"].(string)
		if !ok || reportLink == "" {
			return "", fmt.Errorf("report link is required for report comment email")
		}
		commenterName, ok := data.TemplateData["CommenterName"].(string)
		if !ok || commenterName == "" {
			return "", fmt.Errorf("commenter name is required for report comment email")
		}
		userEmail, ok := data.TemplateData["UserEmail"].(string)
		if !ok || userEmail == "" {
			return "", fmt.Errorf("user email is required for report comment email")
		}

		templateHTML = data.BodyTempate
		templateData = struct {
			UserName      string
			UserEmail     string
			ReportTitle   string
			ReportLink    string
			CommenterName string
		}{
			UserName:      data.RecipientName,
			ReportTitle:   reportTitle,
			UserEmail:     data.To,
			ReportLink:    reportLink,
			CommenterName: commenterName,
		}

	default:
		return "", fmt.Errorf("unsupported email type: %s", data.EmailType)
	}

	tmpl, err := template.New(string(data.EmailType)).Parse(templateHTML)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func SendEmail(data EmailData) error {
	if data.DisabledNotifications {
		return nil
	}

	if data.To == "" || data.RecipientName == "" {
		return fmt.Errorf("recipient email and name cannot be empty")
	}

	email := EmailEmail()
	password := EmailPassword()

	if email == "" || password == "" {
		return fmt.Errorf("email credentials not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", email)
	m.SetHeader("To", data.To)
	m.SetHeader("Subject", data.Subject)

	body, err := RenderEmailTemplate(data)

	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}
	m.SetBody("text/html", body)
	d := gomail.NewDialer("smtp.gmail.com", 587, email, password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}