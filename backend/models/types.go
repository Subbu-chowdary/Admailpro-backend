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
    // ❌ Removed IP field
    UserID          string `json:"userId" bson:"userId"`
    CampaignID      string `json:"campaignId,omitempty" bson:"campaignId,omitempty"`             // New field
    RecipientListID string `json:"recipientListId,omitempty" bson:"recipientListId,omitempty"` // New field
    Status          string `json:"status" bson:"status"`
}

// ❌ Removed IPPair struct and replaced with a simpler Subdomain struct
type Subdomain struct {
    ID        string  `json:"id" bson:"_id"` // A unique ID for the subdomain record
    Name      string  `json:"name" bson:"name"`
    Health    float64 `json:"health" bson:"health"`
    SentCount int64   `json:"sentCount" bson:"sentCount"`
}

// ❌ Removed LinkMapping struct

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