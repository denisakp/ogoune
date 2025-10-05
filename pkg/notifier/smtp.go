package notifier

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"strconv"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
	gomail "gopkg.in/mail.v2"
)

//go:embed templates/*.html
var emailTemplates embed.FS

// SMTPNotifier implements email notifications using SMTP.
type SMTPNotifier struct{}

type TemplateData struct {
	Incident domain.Incident
	Duration string
}

// NewSMTPNotifier creates a new SMTP notifier instance.
func NewSMTPNotifier() *SMTPNotifier {
	return &SMTPNotifier{}
}

// Send sends an email notification for the incident.
// It generates distinct email templates for "Resource Down" and "Resource Up" events.
// The config must contain: recipient, sender, smtp_host, smtp_port, smtp_user, smtp_password
func (n *SMTPNotifier) Send(ctx context.Context, config domain.Integration, incident domain.Incident) error {
	// Extract configuration from Integration Config field
	configMap, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// Extract required fields
	recipient, ok := configMap["recipient"].(string)
	if !ok || recipient == "" {
		return fmt.Errorf("config missing 'recipient' field")
	}

	sender, ok := configMap["sender"].(string)
	if !ok || sender == "" {
		return fmt.Errorf("config missing 'sender' field")
	}

	smtpHost, ok := configMap["smtp_host"].(string)
	if !ok || smtpHost == "" {
		return fmt.Errorf("config missing 'smtp_host' field")
	}

	smtpPortStr, ok := configMap["smtp_port"].(string)
	if !ok || smtpPortStr == "" {
		return fmt.Errorf("config missing 'smtp_port' field")
	}

	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
		return fmt.Errorf("invalid smtp_port: %w", err)
	}

	smtpUser, ok := configMap["smtp_user"].(string)
	if !ok || smtpUser == "" {
		return fmt.Errorf("config missing 'smtp_user' field")
	}

	smtpPassword, ok := configMap["smtp_password"].(string)
	if !ok || smtpPassword == "" {
		return fmt.Errorf("config missing 'smtp_password' field")
	}

	// Determine if this is a DOWN or UP notification based on ResolvedAt
	isResolved := incident.ResolvedAt != nil

	var subject string
	var htmlBody string

	if isResolved {
		// Resource is back UP
		subject = fmt.Sprintf("✅ RESOLVED: %s is back online", incident.Resource.Name)
		htmlBody = n.generateUpEmailHTML(incident)
	} else {
		// Resource is DOWN
		subject = fmt.Sprintf("🔴 ALERT: %s is down", incident.Resource.Name)
		htmlBody = n.generateDownEmailHTML(incident)
	}

	// Create email message
	message := gomail.NewMessage()
	message.SetHeader("From", sender)
	message.SetHeader("To", recipient)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", htmlBody)

	// Create SMTP dialer
	dialer := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPassword)

	// Send the email
	if err := dialer.DialAndSend(message); err != nil {
		log.Printf("[SMTP Notifier] Failed to send email to %s: %v", recipient, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("[SMTP Notifier] Successfully sent email to %s\nSubject: %s", recipient, subject)
	return nil
}

// generateDownEmailHTML creates an HTML email for resource down events.
func (n *SMTPNotifier) generateDownEmailHTML(incident domain.Incident) string {
	data := &TemplateData{Incident: incident}

	tmpl, err := template.ParseFS(emailTemplates, "templates/resource_down.html")
	if err != nil {
		log.Printf("[SMTP Notifier] Failed to parse template: %s", err)
		return fmt.Sprintf("<!DOCTYPE html><html><body><p>Resource %s is down.</p></body></html>", incident.Resource.Name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("[SMTP Notifier] Failed to execute template: %s", err)
		return fmt.Sprintf("error generating email content for resource %s", incident.Resource.Name)
	}

	return buf.String()
}

// generateUpEmailHTML creates an HTML email for resource recovery events.
func (n *SMTPNotifier) generateUpEmailHTML(incident domain.Incident) string {
	duration := "N/A"
	if incident.ResolvedAt != nil {
		d := incident.ResolvedAt.Sub(incident.StartedAt)
		duration = formatDuration(d)
	}

	data := TemplateData{Incident: incident, Duration: duration}

	tmpl, err := template.ParseFS(emailTemplates, "templates/resource_up.html")
	if err != nil {
		log.Printf("[SMTP Notifier] Failed to parse template: %s", err)
		return fmt.Sprintf("<!DOCTYPE html><html><body><p>Resource %s is back online.</p></body></html>", incident.Resource.Name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("[SMTP Notifier] Failed to execute template: %s", err)
		return fmt.Sprintf("error generating email content for resource %s", incident.Resource.Name)
	}

	return buf.String()
}

// formatDuration formats a duration into a human-readable string.
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
