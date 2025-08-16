package api

import (
	"email-sender/backend/db"
	"email-sender/backend/models"
	"email-sender/backend/utils"
	"encoding/json"
	"log"
	"time" // Import time for campaign creation timestamp

	"github.com/valyala/fasthttp"
	// Import mongo to check for no documents found
)

// --- Subdomain Management (CRUD) ---

// GetSubdomains handles fetching all IP-subdomain pairs.
func GetSubdomains(ctx *fasthttp.RequestCtx) {
	subdomains, err := db.GetSubdomains()
	if err != nil {
		ctx.Error("Failed to fetch subdomains", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(subdomains)
}

// AddSubdomain handles adding a new subdomain.
func AddSubdomain(ctx *fasthttp.RequestCtx) {
	var subdomain models.Subdomain
	if err := json.Unmarshal(ctx.PostBody(), &subdomain); err != nil {
		ctx.Error("Invalid request", fasthttp.StatusBadRequest)
		return
	}
	if subdomain.Name == "" {
		ctx.Error("Subdomain name is required", fasthttp.StatusBadRequest)
		return
	}

	subdomain.ID = utils.GenerateID()
	if err := db.SaveSubdomain(&subdomain); err != nil {
		ctx.Error("Failed to add subdomain", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusCreated)
}

// UpdateSubdomain handles updating an existing subdomain.
// This handler now correctly uses the ID from the URL parameter.
func UpdateSubdomain(ctx *fasthttp.RequestCtx) {
	subdomainID := ctx.UserValue("id").(string)
	var subdomain models.Subdomain
	if err := json.Unmarshal(ctx.PostBody(), &subdomain); err != nil {
		ctx.Error("Invalid request", fasthttp.StatusBadRequest)
		return
	}
	if err := db.UpdateSubdomain(subdomainID, &subdomain); err != nil {
		ctx.Error("Failed to update subdomain", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
}

// DeleteSubdomain handles deleting a subdomain.
// This handler now correctly uses the ID from the URL parameter.
func DeleteSubdomain(ctx *fasthttp.RequestCtx) {
	subdomainID := ctx.UserValue("id").(string)
	if err := db.DeleteSubdomain(subdomainID); err != nil {
		ctx.Error("Failed to delete subdomain", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}


// --- Campaign Management (CRUD) ---

// CreateCampaignHandler handles the creation of a new email campaign.
func CreateCampaignHandler(ctx *fasthttp.RequestCtx) {
	var req models.Campaign
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

// GetCampaignsHandler handles fetching all campaigns.
func GetCampaignsHandler(ctx *fasthttp.RequestCtx) {
	campaigns, err := db.GetCampaigns()
	if err != nil {
		ctx.Error("Failed to fetch campaigns", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(campaigns)
}

// GetCampaignHandler handles fetching a single campaign by ID.
func GetCampaignHandler(ctx *fasthttp.RequestCtx) {
	campaignID := ctx.UserValue("id").(string)
	campaign, err := db.FindCampaign(campaignID)
	if err != nil {
		ctx.Error("Failed to fetch campaign", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(campaign)
}

// UpdateCampaignHandler handles updating an existing campaign.
func UpdateCampaignHandler(ctx *fasthttp.RequestCtx) {
	campaignID := ctx.UserValue("id").(string)
	var req models.Campaign
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error("Invalid request", fasthttp.StatusBadRequest)
		return
	}
	if err := db.UpdateCampaign(campaignID, &req); err != nil {
		ctx.Error("Failed to update campaign", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
}

// DeleteCampaignHandler handles deleting an existing campaign.
func DeleteCampaignHandler(ctx *fasthttp.RequestCtx) {
	campaignID := ctx.UserValue("id").(string)
	if err := db.DeleteCampaign(campaignID); err != nil {
		ctx.Error("Failed to delete campaign", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}


// --- RecipientList Management (CRUD) ---

// GetRecipientListsHandler handles fetching all recipient lists.
func GetRecipientListsHandler(ctx *fasthttp.RequestCtx) {
	lists, err := db.GetRecipientLists()
	if err != nil {
		ctx.Error("Failed to fetch recipient lists", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(lists)
}

// GetRecipientListHandler handles fetching a single recipient list by ID.
func GetRecipientListHandler(ctx *fasthttp.RequestCtx) {
	listID := ctx.UserValue("id").(string)
	list, err := db.FindRecipientList(listID)
	if err != nil {
		ctx.Error("Failed to fetch recipient list", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(list)
}

// UpdateRecipientListHandler handles updating an existing recipient list.
func UpdateRecipientListHandler(ctx *fasthttp.RequestCtx) {
	listID := ctx.UserValue("id").(string)
	var req models.RecipientList
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		ctx.Error("Invalid request", fasthttp.StatusBadRequest)
		return
	}
	if err := db.UpdateRecipientList(listID, &req); err != nil {
		ctx.Error("Failed to update recipient list", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusOK)
}

// DeleteRecipientListHandler handles deleting an existing recipient list.
func DeleteRecipientListHandler(ctx *fasthttp.RequestCtx) {
	listID := ctx.UserValue("id").(string)
	if err := db.DeleteRecipientList(listID); err != nil {
		ctx.Error("Failed to delete recipient list", fasthttp.StatusInternalServerError)
		return
	}
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

