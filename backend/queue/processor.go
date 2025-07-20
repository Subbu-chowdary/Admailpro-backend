package queue

import (
	"context"
	"email-sender/backend/config"
	"email-sender/backend/db"
	"email-sender/backend/health"
	"email-sender/backend/models"
	"email-sender/backend/smtp"
	"encoding/json"
	"fmt"
	"log"

	// "sync/atomic" // No longer needed as roundRobinIndex is removed

	"github.com/hibiken/asynq" // Correct import for asynq
)


func StartWorker() {
	cfg := config.GetConfig()
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{Concurrency: 10},
	)
	mux := asynq.NewServeMux()
	mux.HandleFunc("email:send", processEmailTask)
	if err := srv.Run(mux); err != nil {
		log.Fatalf("Could not run worker: %v", err)
	}
}

func processEmailTask(ctx context.Context, t *asynq.Task) error {
	var job models.EmailJob
	if err := json.Unmarshal(t.Payload(), &job); err != nil {
		log.Printf("Failed to unmarshal email job payload: %v", err)
		return fmt.Errorf("invalid payload: %w", err)
	}

	// Initialize HealthManager. It will load pairs from DB or generate defaults if DB is empty.
	hm := health.NewHealthManager()

	// Get the healthiest IP/subdomain pair directly from the HealthManager
	selectedPair := hm.GetHealthiestPair()

	if selectedPair.Subdomain == "" || selectedPair.IP == "" {
		// This case implies no IP-subdomain pairs are available in DB or defaults.
		// This should ideally be prevented by ensuring initial setup.
		log.Printf("No healthy IP/subdomain pair available for sending email to %s", job.Request.Recipient)
		return fmt.Errorf("no healthy IP/subdomain pair available")
	}

	job.Subdomain = selectedPair.Subdomain
	job.IP = selectedPair.IP

	log.Printf("Attempting to send email to %s via %s (IP: %s) for CampaignID: %s, RecipientListID: %s",
		job.Request.Recipient, job.Subdomain, job.IP, job.CampaignID, job.RecipientListID)

	// Send the email
	if err := smtp.SendEmail(job); err != nil {
		log.Printf("Failed to send email to %s via %s (IP: %s): %v", job.Request.Recipient, job.Subdomain, job.IP, err)
		// In a real system, you'd analyze the SMTP error to determine
		// if it's a bounce/complaint/spam and update health manager accordingly.
		// For now, let's apply a small health penalty on failure.
		hm.UpdateHealth(selectedPair.Subdomain, selectedPair.IP, 0.05, 0.05, 0.0) // Example: small penalty for send failure
		return fmt.Errorf("failed to send email: %w", err) // Return error to Asynq for retry logic
	}

	// If sending was successful, update health (e.g., increment sent count, no penalty)
	hm.UpdateHealth(selectedPair.Subdomain, selectedPair.IP, 0.0, 0.0, 0.0) // No bounce/spam/complaint for successful send

	// Update job status in DB
	if err := db.GetCollection("email_jobs").FindOneAndUpdate(
		context.Background(),
		map[string]string{"_id": job.ID},
		map[string]interface{}{"$set": map[string]string{"status": "sent"}},
	).Err(); err != nil {
		log.Printf("Failed to update job status for %s: %v", job.ID, err)
		return fmt.Errorf("failed to update job status: %w", err)
	}

	log.Printf("Successfully processed email job %s for %s", job.ID, job.Request.Recipient)
	return nil
}