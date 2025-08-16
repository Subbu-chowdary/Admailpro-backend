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

// SendEmailHandler sends a single email (not part of a campaign) using SES Configuration Set.
func SendEmailHandler(ctx *fasthttp.RequestCtx) {
	var req models.EmailRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error("Invalid request", fasthttp.StatusBadRequest)
		return
	}

	defaultCfg := config.GetConfig()
	selectedConfigSet := defaultCfg.DefaultConfigSet // e.g., "admailpro-configset-1"
	selectedRegion := defaultCfg.DefaultSESRegion    // e.g., "us-east-1"

	job := models.EmailJob{
		ID:        utils.GenerateID(),
		Request:   req,
		Status:    "queued",
		Subdomain: "",
		ConfigSet: selectedConfigSet,
		Region:    selectedRegion,
	}
	if email, ok := ctx.UserValue("email").(string); ok {
		job.UserID = email
	}

	if err := db.SaveEmailJob(&job); err != nil {
		ctx.Error("Failed to save job", fasthttp.StatusInternalServerError)
		return
	}

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: defaultCfg.RedisAddr})
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

// SendCampaignRequest defines the body for SendCampaignHandler.
type SendCampaignRequest struct {
	CampaignID      string `json:"campaign_id"`
	RecipientListID string `json:"recipient_list_id"`
}

// CreateCampaignHandler is a placeholder for creating campaigns.
func CreateCampaignHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString("Campaign created successfully (placeholder)")
}

// SendCampaignHandler queues a campaign using SES Configuration Sets.
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

	if campaign.UserID != userID || recipientList.UserID != userID {
		ctx.Error("Unauthorized: You do not own this campaign or recipient list", fasthttp.StatusForbidden)
		return
	}

	defaultCfg := config.GetConfig()
	batchConfigSet := defaultCfg.DefaultConfigSet
	batchRegion := defaultCfg.DefaultSESRegion
	batchSubdomain := ""

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: defaultCfg.RedisAddr})
	defer client.Close()

	successCount := 0
	for _, recipientEmail := range recipientList.Emails {
		emailRequest := models.EmailRequest{
			Recipient: recipientEmail,
			Subject:   campaign.Subject,
			HTML:      campaign.HTML,
		}

		job := models.EmailJob{
			ID:              utils.GenerateID(),
			Request:         emailRequest,
			Status:          "queued",
			Subdomain:       batchSubdomain,
			ConfigSet:       batchConfigSet,
			Region:          batchRegion,
			UserID:          userID,
			CampaignID:      campaign.ID,
			RecipientListID: recipientList.ID,
		}

		if err := db.SaveEmailJob(&job); err != nil {
			log.Printf("Failed to save email job for %s: %v", recipientEmail, err)
			continue
		}

		payload, _ := json.Marshal(job)
		task := asynq.NewTask("email:send", payload, asynq.MaxRetry(3))
		if _, err := client.Enqueue(task); err != nil {
			log.Printf("Failed to enqueue email job for %s: %v", recipientEmail, err)
			continue
		}
		successCount++
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	response := fmt.Sprintf("Successfully queued %d emails for campaign ID: %s", successCount, campaign.ID)
	ctx.SetBodyString(response)
	log.Println(response)
}
