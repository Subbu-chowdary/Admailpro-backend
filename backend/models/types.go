// backend/models/types.go
package models

type User struct {
	ID       string `json:"id" bson:"_id"`
	Email    string `json:"email" bson:"email"`
	Password string `json:"password" bson:"password"`
}

type EmailRequest struct {
	Recipient string `json:"recipient" bson:"recipient"`
	Subject   string `json:"subject" bson:"subject"`
	HTML      string `json:"html" bson:"html"`
}

type EmailJob struct {
	ID              string `json:"id" bson:"_id"`
	Request         EmailRequest `json:"request" bson:"request"`
	Subdomain       string `json:"subdomain" bson:"subdomain"`
	IP              string `json:"ip" bson:"ip"`
	UserID          string `json:"userId" bson:"userId"`
	CampaignID      string `json:"campaignId,omitempty" bson:"campaignId,omitempty"`             // New field
	RecipientListID string `json:"recipientListId,omitempty" bson:"recipientListId,omitempty"` // New field
	Status          string `json:"status" bson:"status"`
}

type IPPair struct {
	Subdomain string  `json:"subdomain" bson:"subdomain"`
	IP        string  `json:"ip" bson:"ip"`
	Health    float64 `json:"health" bson:"health"`
	SentCount int64   `json:"sentCount" bson:"sentCount"`
}

type LinkMapping struct {
	ID          string `json:"id" bson:"_id"`
	OriginalURL string `json:"originalUrl" bson:"originalUrl"`
}

// New Campaign Model
type Campaign struct {
	ID      string `json:"id" bson:"_id"`
	UserID  string `json:"userId" bson:"userId"` // To link campaign to a user
	Subject string `json:"subject" bson:"subject"`
	HTML    string `json:"html" bson:"html"`
	Created int64  `json:"created" bson:"created"` // Timestamp
}

// New RecipientList Model
type RecipientList struct {
	ID      string   `json:"id" bson:"_id"`
	UserID  string   `json:"userId" bson:"userId"`
	Name    string   `json:"name" bson:"name"` // Optional: Name for the list
	Emails  []string `json:"emails" bson:"emails"`
	Created int64    `json:"created" bson:"created"`
}