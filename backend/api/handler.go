// backend/api/handler.go
package api

import (
	"email-sender/backend/cloaker"
	"email-sender/backend/config"
	"email-sender/backend/db"
	"email-sender/backend/models"
	"email-sender/backend/utils"
	"encoding/json"
	"fmt"
	"log"
	"time" // Import time for campaign creation timestamp

	"github.com/hibiken/asynq"
	"github.com/valyala/fasthttp"
)

// SendEmailHandler for sending a single email directly (not part of campaign flow)
func SendEmailHandler(ctx *fasthttp.RequestCtx) {
	var req models.EmailRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error("Invalid request", fasthttp.StatusBadRequest)
		return
	}

	cloakedHTML, _, err := cloaker.CloakLinks(req.HTML)
	if err != nil {
		ctx.Error("Failed to cloak links", fasthttp.StatusInternalServerError)
		return
	}
	req.HTML = cloakedHTML

	job := models.EmailJob{
		ID:        utils.GenerateID(),
		Request:   req,
		Status:    "queued",
		Subdomain: "", IP: "",
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
		ctx.Error("Failed to enqueue email", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusAccepted)
}

// GetSubdomains handles fetching all IP-subdomain pairs.
func GetSubdomains(ctx *fasthttp.RequestCtx) {
	pairs, err := db.GetIPPairs()
	if err != nil {
		ctx.Error("Failed to fetch subdomains", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(pairs)
}

// AddSubdomain handles adding a new IP-subdomain pair.
func AddSubdomain(ctx *fasthttp.RequestCtx) {
	var pair models.IPPair
	if err := json.Unmarshal(ctx.PostBody(), &pair); err != nil {
		ctx.Error("Invalid request", fasthttp.StatusBadRequest)
		return
	}
	if err := db.SaveIPPair(&pair); err != nil {
		ctx.Error("Failed to add subdomain", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusCreated)
}

// UpdateSubdomain handles updating an existing IP-subdomain pair.
// Note: This implementation assumes subdomain and IP are extracted from UserValue,
// which typically comes from URL parameters via a router.
func UpdateSubdomain(ctx *fasthttp.RequestCtx) {
	subdomain := ctx.UserValue("subdomain").(string)
	ip := ctx.UserValue("ip").(string)
	var pair models.IPPair
	if err := json.Unmarshal(ctx.PostBody(), &pair); err != nil {
		ctx.Error("Invalid request", fasthttp.StatusBadRequest)
		return
	}
	if err := db.UpdateIPPair(subdomain, ip, &pair); err != nil {
		ctx.Error("Failed to update subdomain", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
}

// DeleteSubdomain handles deleting an IP-subdomain pair.
// Note: This implementation assumes subdomain and IP are extracted from UserValue.
func DeleteSubdomain(ctx *fasthttp.RequestCtx) {
	subdomain := ctx.UserValue("subdomain").(string)
	ip := ctx.UserValue("ip").(string)
	if err := db.DeleteIPPair(subdomain, ip); err != nil {
		ctx.Error("Failed to delete subdomain", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

// CreateCampaignRequest defines the expected payload for creating a campaign.
type CreateCampaignRequest struct {
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

// CreateCampaignHandler handles the creation of a new email campaign (subject and HTML content).
// It returns a unique campaign_id.
func CreateCampaignHandler(ctx *fasthttp.RequestCtx) {
	var req CreateCampaignRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error("Invalid request body", fasthttp.StatusBadRequest)
		return
	}

	if req.Subject == "" || req.HTML == "" {
		ctx.Error("Subject and HTML content are required", fasthttp.StatusBadRequest)
		return
	}

	userID, ok := ctx.UserValue("email").(string)
	if !ok || userID == "" {
		ctx.Error("Unauthorized: User not identified", fasthttp.StatusUnauthorized)
		return
	}

	campaign := models.Campaign{
		ID:      utils.GenerateID(),
		UserID:  userID,
		Subject: req.Subject,
		HTML:    req.HTML,
		Created: time.Now().Unix(),
	}

	if err := db.CreateCampaign(&campaign); err != nil {
		log.Printf("Failed to create campaign: %v", err)
		ctx.Error("Failed to create campaign", fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(map[string]string{"campaign_id": campaign.ID})
}

// SendCampaignRequest defines the payload for initiating a campaign send to a recipient list.
type SendCampaignRequest struct {
	CampaignID      string `json:"campaign_id"`
	RecipientListID string `json:"recipient_list_id"`
}

// SendCampaignHandler retrieves campaign content and recipient list, then enqueues email jobs.
// This is the "send" trigger after both content and recipient list are prepared.
func SendCampaignHandler(ctx *fasthttp.RequestCtx) {
	var req SendCampaignRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error("Invalid request body", fasthttp.StatusBadRequest)
		return
	}

	if req.CampaignID == "" || req.RecipientListID == "" {
		ctx.Error("Campaign ID and Recipient List ID are required", fasthttp.StatusBadRequest)
		return
	}

	userID, ok := ctx.UserValue("email").(string)
	if !ok || userID == "" {
		ctx.Error("Unauthorized: User not identified", fasthttp.StatusUnauthorized)
		return
	}

	// 1. Fetch Campaign (subject and HTML)
	campaign, err := db.FindCampaign(req.CampaignID)
	if err != nil {
		log.Printf("Failed to find campaign %s: %v", req.CampaignID, err)
		ctx.Error("Campaign not found or internal error", fasthttp.StatusNotFound)
		return
	}
	// Security check: Ensure campaign belongs to the authenticated user
	if campaign.UserID != userID {
		ctx.Error("Unauthorized: Campaign does not belong to this user", fasthttp.StatusUnauthorized)
		return
	}

	// 2. Fetch Recipient List
	recipientList, err := db.FindRecipientList(req.RecipientListID)
	if err != nil {
		log.Printf("Failed to find recipient list %s: %v", req.RecipientListID, err)
		ctx.Error("Recipient list not found or internal error", fasthttp.StatusNotFound)
		return
	}
	// Security check: Ensure recipient list belongs to the authenticated user
	if recipientList.UserID != userID {
		ctx.Error("Unauthorized: Recipient list does not belong to this user", fasthttp.StatusUnauthorized)
		return
	}

	if len(recipientList.Emails) == 0 {
		ctx.Error("Recipient list is empty", fasthttp.StatusBadRequest)
		return
	}

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

		// Cloak links for each email
		cloakedHTML, _, err := cloaker.CloakLinks(emailRequest.HTML)
		if err != nil {
			log.Printf("Failed to cloak links for %s (campaign %s): %v", recipientEmail, campaign.ID, err)
			continue // Skip this email but continue processing others
		}
		emailRequest.HTML = cloakedHTML

		job := models.EmailJob{
			ID:              utils.GenerateID(),
			Request:         emailRequest,
			Status:          "queued",
			Subdomain:       "", IP: "", // These will be set by the queue processor
			UserID:          userID,
			CampaignID:      campaign.ID,
			RecipientListID: recipientList.ID,
		}

		if err := db.SaveEmailJob(&job); err != nil {
			log.Printf("Failed to save email job for %s (campaign %s): %v", recipientEmail, campaign.ID, err)
			continue
		}

		payload, _ := json.Marshal(job)
		task := asynq.NewTask("email:send", payload, asynq.MaxRetry(3))
		if _, err := client.Enqueue(task); err != nil {
			log.Printf("Failed to enqueue email job for %s (campaign %s): %v", recipientEmail, campaign.ID, err)
			continue
		}
		successCount++
	}

	ctx.SetStatusCode(fasthttp.StatusAccepted)
	ctx.SetBody([]byte(fmt.Sprintf(
		"Successfully enqueued %d email jobs for campaign '%s' to recipient list '%s'.",
		successCount, campaign.ID, recipientList.ID,
	)))
}