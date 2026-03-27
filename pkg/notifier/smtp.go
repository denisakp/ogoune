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
type SMTPNotifier struct {
	recipient    string
	sender       string
	smtpHost     string
	smtpPort     string
	smtpUser     string
	smtpPassword string
}

type TemplateData struct {
	Incident domain.Incident
	Duration string
}

type ComponentTemplateData struct {
	Name     string
	Status   domain.ComponentStatus
	Impacted []ComponentResource
}

type ExpiryTemplateData struct {
	ResourceName  string
	ResourceURL   string
	ExpiryType    string
	DaysRemaining int
	ExpiresAt     string
	Issuer        string
	Threshold     int
}

type TestTemplateData struct {
	Message   string
	Timestamp string
}

// NewSMTPNotifier creates a new SMTP notifier instance with the provided configuration.
func NewSMTPNotifier(recipient, sender, smtpHost, smtpPort, smtpUser, smtpPassword string) *SMTPNotifier {
	return &SMTPNotifier{
		recipient:    recipient,
		sender:       sender,
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUser:     smtpUser,
		smtpPassword: smtpPassword,
	}
}

func (n *SMTPNotifier) validateConfig() error {
	if n.recipient == "" {
		return fmt.Errorf("smtp recipient is not configured")
	}
	if n.sender == "" {
		return fmt.Errorf("smtp sender is not configured")
	}
	if n.smtpHost == "" {
		return fmt.Errorf("smtp host is not configured")
	}
	if n.smtpPort == "" {
		return fmt.Errorf("smtp port is not configured")
	}
	if n.smtpUser == "" {
		return fmt.Errorf("smtp user is not configured")
	}
	if n.smtpPassword == "" {
		return fmt.Errorf("smtp password is not configured")
	}
	return nil
}

// Send sends an email notification for either a resource incident or a component state change.
func (n *SMTPNotifier) Send(ctx context.Context, payload NotificationPayload) error {
	if err := n.validateConfig(); err != nil {
		return err
	}

	smtpPort, err := strconv.Atoi(n.smtpPort)
	if err != nil {
		return fmt.Errorf("invalid smtp_port: %w", err)
	}

	var subject string
	var htmlBody string

	switch {
	case payload.Flapping != nil:
		subject = n.flappingSubject(payload.Flapping)
		htmlBody = n.generateFlappingEmailHTML(payload.Flapping)
	case payload.Reminder != nil:
		subject = n.reminderSubject(payload.Reminder)
		htmlBody = n.generateReminderEmailHTML(payload.Reminder)
	case payload.Component != nil:
		subject = n.componentSubject(payload.Component)
		htmlBody = n.generateComponentEmailHTML(payload.Component)
	case payload.Expiry != nil:
		subject = n.expirySubject(payload.Expiry)
		htmlBody = n.generateExpiryEmailHTML(payload.Expiry)
	case payload.Incident != nil:
		incident := *payload.Incident
		subject, htmlBody = n.incidentEmailContent(incident)
	default:
		return fmt.Errorf("notification payload missing incident, component, expiry, flapping, or reminder")
	}

	message := gomail.NewMessage()
	message.SetHeader("From", n.sender)
	message.SetHeader("To", n.recipient)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", htmlBody)

	dialer := gomail.NewDialer(n.smtpHost, smtpPort, n.smtpUser, n.smtpPassword)

	if err := dialer.DialAndSend(message); err != nil {
		log.Printf("[SMTP Notifier] Failed to send email to %s: %v", n.recipient, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("[SMTP Notifier] Successfully sent email to %s\nSubject: %s", n.recipient, subject)
	return nil
}

func (n *SMTPNotifier) incidentEmailContent(incident domain.Incident) (string, string) {
	isResolved := incident.ResolvedAt != nil

	if isResolved {
		subject := fmt.Sprintf("✅ RESOLVED: %s is back online", incident.Resource.Name)
		return subject, n.generateUpEmailHTML(incident)
	}

	subject := fmt.Sprintf("🔴 ALERT: %s is down", incident.Resource.Name)
	return subject, n.generateDownEmailHTML(incident)
}

func (n *SMTPNotifier) componentSubject(component *ComponentNotification) string {
	switch component.Status {
	case domain.ComponentStatusDown:
		return fmt.Sprintf("🔴 Component %s is down", component.Component.Name)
	case domain.ComponentStatusDegraded:
		return fmt.Sprintf("⚠️ Component %s degraded", component.Component.Name)
	default:
		return fmt.Sprintf("✅ Component %s recovered", component.Component.Name)
	}
}

// generateDownEmailHTML creates an HTML email for resource down events.
func (n *SMTPNotifier) generateDownEmailHTML(incident domain.Incident) string {
	data := &TemplateData{Incident: incident}

	tmpl, err := template.ParseFS(emailTemplates, "templates/resource_down.html")
	if err != nil {
		log.Printf("[SMTP Notifier] Failed to parse resource_down.html template: %v", err)
		return fmt.Sprintf("<!DOCTYPE html><html><body><p>Resource %s is down.</p></body></html>", incident.Resource.Name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("[SMTP Notifier] Failed to execute resource_down.html template: %v | Resource: %s | Incident ID: %s", err, incident.Resource.Name, incident.ID)
		return fmt.Sprintf("<!DOCTYPE html><html><body><p>Resource %s is down. Incident ID: %s</p></body></html>", incident.Resource.Name, incident.ID)
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
		log.Printf("[SMTP Notifier] Failed to parse resource_up.html template: %v", err)
		return fmt.Sprintf("<!DOCTYPE html><html><body><p>Resource %s is back online.</p></body></html>", incident.Resource.Name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("[SMTP Notifier] Failed to execute resource_up.html template: %v | Resource: %s | Incident ID: %s", err, incident.Resource.Name, incident.ID)
		return fmt.Sprintf("<!DOCTYPE html><html><body><p>Resource %s is back online. Downtime: %s</p></body></html>", incident.Resource.Name, duration)
	}

	return buf.String()
}

// generateComponentEmailHTML creates an HTML email for component state changes.
func (n *SMTPNotifier) generateComponentEmailHTML(component *ComponentNotification) string {
	data := &ComponentTemplateData{
		Name:     component.Component.Name,
		Status:   component.Status,
		Impacted: component.Impacted,
	}

	tmpl, err := template.ParseFS(emailTemplates, "templates/component_status.html")
	if err != nil {
		log.Printf("[SMTP Notifier] Failed to parse component_status.html template: %v", err)
		return fmt.Sprintf("<!DOCTYPE html><html><body><p>Component %s is %s.</p></body></html>", component.Component.Name, component.Status)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("[SMTP Notifier] Failed to execute component_status.html template: %v | Component: %s", err, component.Component.Name)
		return fmt.Sprintf("<!DOCTYPE html><html><body><p>Component %s is %s.</p></body></html>", component.Component.Name, component.Status)
	}

	return buf.String()
}

// generateTestEmailHTML creates an HTML email for test notifications.
func (n *SMTPNotifier) generateTestEmailHTML(resource domain.Resource) string {
	data := TestTemplateData{
		Message:   fmt.Sprintf("Test notification for resource: %s (%s)", resource.Name, resource.Target),
		Timestamp: time.Now().Format("2006-01-02 15:04:05 MST"),
	}

	tmpl, err := template.ParseFS(emailTemplates, "templates/test.html")
	if err != nil {
		log.Printf("[SMTP Notifier] Failed to parse test template: %s", err)
		return fmt.Sprintf("<!DOCTYPE html><html><body><p>Test notification for resource %s.</p></body></html>", resource.Name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("[SMTP Notifier] Failed to execute test template: %s", err)
		return fmt.Sprintf("error generating test email content for resource %s", resource.Name)
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

// SendTestNotification sends a test email notification for a resource.
// It uses the test.html template to send a simple test message.
func (n *SMTPNotifier) SendTestNotification(ctx context.Context, resource domain.Resource) error {
	// Validate configuration
	if n.recipient == "" {
		return fmt.Errorf("smtp recipient is not configured")
	}
	if n.sender == "" {
		return fmt.Errorf("smtp sender is not configured")
	}
	if n.smtpHost == "" {
		return fmt.Errorf("smtp host is not configured")
	}
	if n.smtpPort == "" {
		return fmt.Errorf("smtp port is not configured")
	}
	if n.smtpUser == "" {
		return fmt.Errorf("smtp user is not configured")
	}
	if n.smtpPassword == "" {
		return fmt.Errorf("smtp password is not configured")
	}

	smtpPort, err := strconv.Atoi(n.smtpPort)
	if err != nil {
		return fmt.Errorf("invalid smtp_port: %w", err)
	}

	// Generate test email content
	subject := fmt.Sprintf("ℹ️ Test Notification - %s", resource.Name)
	htmlBody := n.generateTestEmailHTML(resource)

	// Create email message
	message := gomail.NewMessage()
	message.SetHeader("From", n.sender)
	message.SetHeader("To", n.recipient)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", htmlBody)

	// Create SMTP dialer
	dialer := gomail.NewDialer(n.smtpHost, smtpPort, n.smtpUser, n.smtpPassword)

	// Send the email
	if err := dialer.DialAndSend(message); err != nil {
		log.Printf("[SMTP Notifier] Failed to send test email to %s: %v", n.recipient, err)
		return fmt.Errorf("failed to send test email: %w", err)
	}

	log.Printf("[SMTP Notifier] Successfully sent test email to %s\nSubject: %s", n.recipient, subject)
	return nil
}

func (n *SMTPNotifier) expirySubject(expiry *ExpiryNotification) string {
	typeLabel := "SSL"
	if expiry.ExpiryType == "domain" {
		typeLabel = "Domain"
	}
	return fmt.Sprintf("%s expiry alert: %s expires in %d days", typeLabel, expiry.Resource.Name, expiry.DaysRemaining)
}

func (n *SMTPNotifier) flappingSubject(flapping *FlappingNotification) string {
	if flapping.Stabilized {
		return fmt.Sprintf("[PulseGuard] %s stabilized after flapping", flapping.Resource.Name)
	}
	return fmt.Sprintf("[PulseGuard] %s is flapping - %d transitions in %d minutes", flapping.Resource.Name, flapping.TransitionCount, flapping.WindowSeconds/60)
}

func (n *SMTPNotifier) generateFlappingEmailHTML(flapping *FlappingNotification) string {
	if flapping.Stabilized {
		return fmt.Sprintf("<!DOCTYPE html><html><body><p>Monitor <strong>%s</strong> stabilized after a flapping episode.</p><p>Current status: %s</p></body></html>", flapping.Resource.Name, flapping.FinalStatus)
	}
	return fmt.Sprintf("<!DOCTYPE html><html><body><p>Your monitor <strong>%s</strong> (%s) has been switching between UP and DOWN %d times in the last %d minutes.</p><p>PulseGuard has suppressed repeated alerts while the service is unstable.</p></body></html>", flapping.Resource.Name, flapping.Resource.Target, flapping.TransitionCount, flapping.WindowSeconds/60)
}

func (n *SMTPNotifier) reminderSubject(reminder *ReminderNotification) string {
	return fmt.Sprintf("[PulseGuard] %s still down - %d minutes", reminder.Resource.Name, reminder.ElapsedMinutes)
}

func (n *SMTPNotifier) generateReminderEmailHTML(reminder *ReminderNotification) string {
	return fmt.Sprintf("<!DOCTYPE html><html><body><p>Your monitor <strong>%s</strong> has been down for %d minutes.</p><p>Incident started at %s.</p><p>This is a reminder - the incident is still active.</p></body></html>", reminder.Resource.Name, reminder.ElapsedMinutes, reminder.Incident.StartedAt.Format(time.RFC3339))
}

func (n *SMTPNotifier) generateExpiryEmailHTML(expiry *ExpiryNotification) string {
	data := &ExpiryTemplateData{
		ResourceName:  expiry.Resource.Name,
		ResourceURL:   expiry.Resource.Target,
		ExpiryType:    expiry.ExpiryType,
		DaysRemaining: expiry.DaysRemaining,
		ExpiresAt:     expiry.ExpiresAt.Format("2006-01-02"),
		Issuer:        expiry.Issuer,
		Threshold:     expiry.Threshold,
	}

	tmpl, err := template.ParseFS(emailTemplates, "templates/expiry_alert.html")
	if err != nil {
		log.Printf("[SMTP Notifier] Failed to parse expiry_alert.html template: %v", err)
		return fmt.Sprintf("<!DOCTYPE html><html><body><p>%s expiry alert: %s expires in %d days.</p></body></html>",
			expiry.ExpiryType, expiry.Resource.Name, expiry.DaysRemaining)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Printf("[SMTP Notifier] Failed to execute expiry_alert.html template: %v | Resource: %s", err, expiry.Resource.Name)
		return fmt.Sprintf("<!DOCTYPE html><html><body><p>%s expiry alert: %s expires in %d days.</p></body></html>",
			expiry.ExpiryType, expiry.Resource.Name, expiry.DaysRemaining)
	}

	return buf.String()
}
