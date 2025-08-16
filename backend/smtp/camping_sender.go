package smtp

import (
	"email-sender/backend/config"
	"email-sender/backend/db"
	"email-sender/backend/models"
	"email-sender/backend/utils"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/valyala/fasthttp"
)

// SendEmailHandler for sending a single email directly (not part of campaign flow)
// This is the original function.
func SendEmailHandler(ctx *fasthttp.RequestCtx) {
	var req models.EmailRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error("Invalid request", fasthttp.StatusBadRequest)
		return
	}

	job := models.EmailJob{
		ID:        utils.GenerateID(),
		Request:   req,
		Status:    "queued",
		Subdomain: "",
		// The IP field is not used here as per your request to rely on Configuration Sets.
	}
	if email, ok := ctx.UserValue("email").(string); ok {
		job.UserID = email
	}
	if err := db.SaveEmailJob(&job); err != nil {
		ctx.Error("Failed to save job", fasthttp.StatusInternalServerError)
		return
	}

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: config.GetConfig().RedisAddr})
	defer client.Close()
	payload, _ := json.Marshal(job)
	task := asynq.NewTask("email:send", payload, asynq.MaxRetry(3))
	if _, err := client.Enqueue(task); err != nil {
		ctx.Error("Failed to enqueue task", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString("Email job queued successfully")
}

// SendCampaignRequest is the request body for the SendCampaignHandler.
type SendCampaignRequest struct {
	CampaignID      string `json:"campaign_id"`
	RecipientListID string `json:"recipient_list_id"`
}

// CreateCampaignHandler is a placeholder for the campaign creation handler.
func CreateCampaignHandler(ctx *fasthttp.RequestCtx) {
	// ... implementation for creating a campaign would go here
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString("Campaign created successfully (placeholder)")
}

// SendCampaignHandler handles sending a campaign to a list of recipients.
func SendCampaignHandler(ctx *fasthttp.RequestCtx) {
	var req SendCampaignRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error("Invalid request", fasthttp.StatusBadRequest)
		return
	}

	userID, ok := ctx.UserValue("email").(string)
	if !ok || userID == "" {
		ctx.Error("Unauthorized: User not identified", fasthttp.StatusUnauthorized)
		return
	}

	// Retrieve the campaign and recipient list from the database
	campaign, err := db.FindCampaign(req.CampaignID)
	if err != nil {
		ctx.Error(fmt.Sprintf("Campaign not found: %v", err), fasthttp.StatusNotFound)
		return
	}

	recipientList, err := db.FindRecipientList(req.RecipientListID)
	if err != nil {
		ctx.Error(fmt.Sprintf("Recipient list not found: %v", err), fasthttp.StatusNotFound)
		return
	}

	// Verify ownership of campaign and recipient list
	if campaign.UserID != userID || recipientList.UserID != userID {
		ctx.Error("Unauthorized: You do not own this campaign or recipient list", fasthttp.StatusForbidden)
		return
	}

	// Initialize the Asynq client
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: config.GetConfig().RedisAddr})
	defer client.Close()

	successCount := 0
	for _, recipientEmail := range recipientList.Emails {
		// Use subject and HTML from the retrieved campaign
		emailRequest := models.EmailRequest{
			Recipient: recipientEmail,
			Subject:   campaign.Subject,
			HTML:      campaign.HTML,
		}

		// Create an email job for the queue
		job := models.EmailJob{
			ID:              utils.GenerateID(),
			Request:         emailRequest,
			Status:          "queued",
			Subdomain:       "",  // These will be set by the queue processor
			UserID:          userID,
			CampaignID:      campaign.ID,
			RecipientListID: recipientList.ID,
		}

		// Save the job to the database before enqueuing
		if err := db.SaveEmailJob(&job); err != nil {
			log.Printf("Failed to save email job for %s (campaign %s): %v", recipientEmail, campaign.ID, err)
			continue
		}

		// Enqueue the job
		payload, _ := json.Marshal(job)
		task := asynq.NewTask("email:send", payload, asynq.MaxRetry(3))
		if _, err := client.Enqueue(task); err != nil {
			log.Printf("Failed to enqueue email job for %s (campaign %s): %v", recipientEmail, campaign.ID, err)
			continue
		}
		successCount++
	}

	// Return a summary of the operation
	ctx.SetStatusCode(fasthttp.StatusOK)
	response := fmt.Sprintf("Successfully queued %d emails for campaign ID: %s", successCount, campaign.ID)
	ctx.SetBodyString(response)
	log.Println(response)
}
