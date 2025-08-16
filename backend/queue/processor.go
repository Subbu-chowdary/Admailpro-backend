package queue

import (
	"context"
	"email-sender/backend/config"
	"email-sender/backend/db"
	"email-sender/backend/models"
	"email-sender/backend/smtp"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq" // Correct import for asynq
)

// StartWorker initializes and starts the Asynq worker.
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

// processEmailTask handles the email sending logic for a single task.
func processEmailTask(ctx context.Context, t *asynq.Task) error {
    var job models.EmailJob
    if err := json.Unmarshal(t.Payload(), &job); err != nil {
        log.Printf("Failed to unmarshal email job payload: %v", err)
        return fmt.Errorf("invalid payload: %w", err)
    }

    log.Printf("Attempting to send email to %s for CampaignID: %s, RecipientListID: %s",
        job.Request.Recipient, job.CampaignID, job.RecipientListID)

    // Send the email
    if err := smtp.SendEmail(job); err != nil {
        log.Printf("Failed to send email to %s: %v", job.Request.Recipient, err)
        return fmt.Errorf("failed to send email: %w", err)
    }

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