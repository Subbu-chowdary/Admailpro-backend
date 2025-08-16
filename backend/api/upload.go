// backend/api/upload.go
package api

import (
	"email-sender/backend/db"
	"email-sender/backend/models"
	"email-sender/backend/utils"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time" // Import time for timestamp

	"github.com/valyala/fasthttp"
)

// UploadCSVHandler handles the upload of a CSV file containing only recipient emails.
// It saves the list of emails and returns a recipient_list_id.
// It does NOT enqueue email jobs at this stage.
func UploadCSVHandler(ctx *fasthttp.RequestCtx) {
	contentType := string(ctx.Request.Header.ContentType())
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		ctx.Error("Invalid content type", fasthttp.StatusBadRequest)
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.Error("Failed to parse multipart form", fasthttp.StatusBadRequest)
		return
	}

	files := form.File["file"]
	if len(files) == 0 {
		ctx.Error("No CSV file provided", fasthttp.StatusBadRequest)
		return
	}

	// Optional: Get a name for the recipient list from form data
	listName := "Untitled List"
	if names := form.Value["list_name"]; len(names) > 0 {
		listName = string(names[0])
	}

	file := files[0]
	src, err := file.Open()
	if err != nil {
		ctx.Error("Failed to open uploaded file", fasthttp.StatusInternalServerError)
		return
	}
	defer src.Close()

	reader := csv.NewReader(src)
	records, err := reader.ReadAll()
	if err != nil {
		ctx.Error("Invalid CSV format", fasthttp.StatusBadRequest)
		return
	}

	var recipients []string
	if len(records) < 2 { // Require at least a header and one data row
		ctx.Error("CSV must contain at least a header and one recipient", fasthttp.StatusBadRequest)
		return
	}

	// Iterate from the second row (skip header)
	for i, record := range records {
		if i == 0 { // Skip header row
			continue
		}

		if len(record) == 0 {
			log.Printf("Skipping empty row %d in CSV", i+1)
			continue
		}

		// Assume first column is the email address
		email := strings.TrimSpace(record[0])
		if email == "" {
			log.Printf("Skipping row %d: Email address is empty", i+1)
			continue
		}
		recipients = append(recipients, email)
	}

	if len(recipients) == 0 {
		ctx.Error("No valid recipient emails found in the CSV", fasthttp.StatusBadRequest)
		return
	}

	userID, ok := ctx.UserValue("email").(string)
	if !ok || userID == "" {
		ctx.Error("Unauthorized: User not identified", fasthttp.StatusUnauthorized)
		return
	}

	recipientList := models.RecipientList{
		ID:      utils.GenerateID(),
		UserID:  userID,
		Name:    listName,
		Emails:  recipients,
		Created: time.Now().Unix(),
	}

	if err := db.CreateRecipientList(&recipientList); err != nil {
		log.Printf("Failed to save recipient list: %v", err)
		ctx.Error("Failed to save recipient list", fasthttp.StatusInternalServerError)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusCreated)
	ctx.SetContentType("application/json")
	json.NewEncoder(ctx).Encode(map[string]string{
		"message":           fmt.Sprintf("Successfully uploaded %d recipients to list '%s'", len(recipients), listName),
		"recipient_list_id": recipientList.ID,
	})
}