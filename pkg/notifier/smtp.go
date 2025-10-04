package notifier

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/denisakp/pulseguard/internal/domain"
)

// SMTPNotifier implements email notifications using SMTP.
type SMTPNotifier struct{}

// NewSMTPNotifier creates a new SMTP notifier instance.
func NewSMTPNotifier() *SMTPNotifier {
	return &SMTPNotifier{}
}

// Send sends an email notification for the incident.
// It generates distinct email templates for "Resource Down" and "Resource Up" events.
// For now, this logs the email details. In production, this would use an SMTP library like gomail.
func (n *SMTPNotifier) Send(ctx context.Context, config domain.Integration, incident domain.Incident) error {
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

	// Log the notification (in production, send actual email)
	log.Printf("[SMTP Notifier] Sending email to %s\nSubject: %s\nBody preview: %d bytes\n",
		config.Target, subject, len(htmlBody))

	// TODO: Implement actual SMTP sending with gomail or similar
	// Example:
	// m := gomail.NewMessage()
	// m.SetHeader("From", "alerts@pulseguard.com")
	// m.SetHeader("To", config.Target)
	// m.SetHeader("Subject", subject)
	// m.SetBody("text/html", htmlBody)
	// d := gomail.NewDialer("smtp.example.com", 587, "user", "pass")
	// return d.DialAndSend(m)

	return nil
}

// generateDownEmailHTML creates an HTML email for resource down events.
func (n *SMTPNotifier) generateDownEmailHTML(incident domain.Incident) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background-color: #d32f2f; color: white; padding: 20px; border-radius: 5px 5px 0 0; }
		.content { background-color: #f9f9f9; padding: 20px; border: 1px solid #ddd; border-top: none; }
		.detail { margin: 10px 0; }
		.label { font-weight: bold; color: #555; }
		.value { color: #333; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #777; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h2>🔴 Resource Down Alert</h2>
		</div>
		<div class="content">
			<p>A critical resource is currently unavailable:</p>
			
			<div class="detail">
				<span class="label">Resource:</span>
				<span class="value">%s</span>
			</div>
			
			<div class="detail">
				<span class="label">Incident ID:</span>
				<span class="value">%s</span>
			</div>
			
			<div class="detail">
				<span class="label">Cause:</span>
				<span class="value">%s</span>
			</div>
			
			<div class="detail">
				<span class="label">Started At:</span>
				<span class="value">%s</span>
			</div>
			
			<div class="detail">
				<span class="label">Target:</span>
				<span class="value">%s</span>
			</div>
			
			<p style="margin-top: 20px;">
				<strong>Action Required:</strong> Please investigate this incident immediately.
			</p>
		</div>
		<div class="footer">
			<p>This is an automated alert from Pulseguard monitoring system.</p>
		</div>
	</div>
</body>
</html>
`,
		incident.Resource.Name,
		incident.ID,
		incident.Cause,
		incident.StartedAt.Format("2006-01-02 15:04:05 MST"),
		incident.Resource.Target,
	)
}

// generateUpEmailHTML creates an HTML email for resource recovery events.
func (n *SMTPNotifier) generateUpEmailHTML(incident domain.Incident) string {
	duration := "N/A"
	if incident.ResolvedAt != nil {
		d := incident.ResolvedAt.Sub(incident.StartedAt)
		duration = formatDuration(d)
	}

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<style>
		body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
		.container { max-width: 600px; margin: 0 auto; padding: 20px; }
		.header { background-color: #388e3c; color: white; padding: 20px; border-radius: 5px 5px 0 0; }
		.content { background-color: #f9f9f9; padding: 20px; border: 1px solid #ddd; border-top: none; }
		.detail { margin: 10px 0; }
		.label { font-weight: bold; color: #555; }
		.value { color: #333; }
		.footer { margin-top: 20px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #777; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h2>✅ Resource Recovered</h2>
		</div>
		<div class="content">
			<p>Good news! The resource is back online:</p>
			
			<div class="detail">
				<span class="label">Resource:</span>
				<span class="value">%s</span>
			</div>
			
			<div class="detail">
				<span class="label">Incident ID:</span>
				<span class="value">%s</span>
			</div>
			
			<div class="detail">
				<span class="label">Original Cause:</span>
				<span class="value">%s</span>
			</div>
			
			<div class="detail">
				<span class="label">Downtime Duration:</span>
				<span class="value">%s</span>
			</div>
			
			<div class="detail">
				<span class="label">Resolved At:</span>
				<span class="value">%s</span>
			</div>
			
			<div class="detail">
				<span class="label">Target:</span>
				<span class="value">%s</span>
			</div>
			
			<p style="margin-top: 20px; color: #388e3c;">
				<strong>Status:</strong> The resource is now operational.
			</p>
		</div>
		<div class="footer">
			<p>This is an automated notification from Pulseguard monitoring system.</p>
		</div>
	</div>
</body>
</html>
`,
		incident.Resource.Name,
		incident.ID,
		incident.Cause,
		duration,
		incident.ResolvedAt.Format("2006-01-02 15:04:05 MST"),
		incident.Resource.Target,
	)
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

// getIncidentStatus returns a human-readable status string.
func getIncidentStatus(incident domain.Incident) string {
	if incident.ResolvedAt != nil {
		return "Resolved"
	}
	return "Active"
}
