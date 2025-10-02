package notifier

import (
	"context"
	"fmt"
	"log"

	"github.com/denisakp/pulseguard/internal/domain"
)

// SMTPNotifier implements email notifications using SMTP.
type SMTPNotifier struct{}

// NewSMTPNotifier creates a new SMTP notifier instance.
func NewSMTPNotifier() *SMTPNotifier {
	return &SMTPNotifier{}
}

// Send sends an email notification for the incident.
// For now, this logs the email details. In production, this would use an SMTP library like gomail.
func (n *SMTPNotifier) Send(ctx context.Context, config domain.Integration, incident domain.Incident) error {
	// In production, this would:
	// 1. Parse config.Target as email addresses (could be comma-separated)
	// 2. Connect to SMTP server (credentials from env or config)
	// 3. Compose email with incident details
	// 4. Send email

	subject := fmt.Sprintf("Alert: Incident %s for Resource %s", incident.ID, incident.ResourceID)
	body := fmt.Sprintf(`
Incident Alert

Incident ID: %s
Resource ID: %s
Status: %s
Reason: %s
Started At: %s
Resolved: %v

Details: %s
`, incident.ID, incident.ResourceID,
		getIncidentStatus(incident),
		incident.Reason,
		incident.StartedAt.Format("2006-01-02 15:04:05"),
		incident.IsResolved,
		string(incident.Details))

	// Log the notification (in production, send actual email)
	log.Printf("[SMTP Notifier] Sending email to %s\nSubject: %s\nBody: %s\n",
		config.Target, subject, body)

	// TODO: Implement actual SMTP sending with gomail or similar
	// Example:
	// m := gomail.NewMessage()
	// m.SetHeader("From", "alerts@pulseguard.com")
	// m.SetHeader("To", config.Target)
	// m.SetHeader("Subject", subject)
	// m.SetBody("text/plain", body)
	// d := gomail.NewDialer("smtp.example.com", 587, "user", "pass")
	// return d.DialAndSend(m)

	return nil
}

func getIncidentStatus(incident domain.Incident) string {
	if incident.IsResolved {
		return "Resolved"
	}
	return "Active"
}
