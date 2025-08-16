// backend/db/mongo.go
package db

import (
	"context"
	"email-sender/backend/config"
	"email-sender/backend/models"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func init() {
	cfg := config.GetConfig()
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)

	var err error
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("MongoDB ping failed:", err)
	}
	log.Println("✅ Connected to MongoDB")
}

func GetCollection(name string) *mongo.Collection {
	return client.Database("email_sender").Collection(name)
}

// --- User Management ---
func CreateUser(user *models.User) error {
	_, err := GetCollection("users").InsertOne(context.Background(), user)
	return err
}

func FindUser(email string) (models.User, error) {
	var user models.User
	err := GetCollection("users").FindOne(context.Background(), map[string]string{"email": email}).Decode(&user)
	return user, err
}

// --- Email Job Management ---
func SaveEmailJob(job *models.EmailJob) error {
	_, err := GetCollection("email_jobs").InsertOne(context.Background(), job)
	return err
}

// --- Subdomain Management (CRUD) ---
func SaveSubdomain(subdomain *models.Subdomain) error {
	_, err := GetCollection("subdomains").InsertOne(context.Background(), subdomain)
	return err
}

func GetSubdomains() ([]models.Subdomain, error) {
	cursor, err := GetCollection("subdomains").Find(context.Background(), map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	var subdomains []models.Subdomain
	if err = cursor.All(context.Background(), &subdomains); err != nil {
		return nil, err
	}
	return subdomains, nil
}

func UpdateSubdomain(subdomainID string, subdomain *models.Subdomain) error {
	_, err := GetCollection("subdomains").UpdateOne(
		context.Background(),
		bson.M{"_id": subdomainID},
		bson.M{"$set": subdomain},
	)
	return err
}

func DeleteSubdomain(subdomainID string) error {
	_, err := GetCollection("subdomains").DeleteOne(
		context.Background(),
		bson.M{"_id": subdomainID},
	)
	return err
}

// --- Campaign Management (CRUD) ---
func CreateCampaign(campaign *models.Campaign) error {
	_, err := GetCollection("campaigns").InsertOne(context.Background(), campaign)
	return err
}

func FindCampaign(id string) (models.Campaign, error) {
	var campaign models.Campaign
	err := GetCollection("campaigns").FindOne(context.Background(), bson.M{"_id": id}).Decode(&campaign)
	return campaign, err
}

func GetCampaigns() ([]models.Campaign, error) {
	cursor, err := GetCollection("campaigns").Find(context.Background(), map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	var campaigns []models.Campaign
	if err = cursor.All(context.Background(), &campaigns); err != nil {
		return nil, err
	}
	return campaigns, nil
}

func UpdateCampaign(campaignID string, campaign *models.Campaign) error {
	_, err := GetCollection("campaigns").UpdateOne(
		context.Background(),
		bson.M{"_id": campaignID},
		bson.M{"$set": campaign},
	)
	return err
}

func DeleteCampaign(campaignID string) error {
	_, err := GetCollection("campaigns").DeleteOne(
		context.Background(),
		bson.M{"_id": campaignID},
	)
	return err
}

// --- RecipientList Management (CRUD) ---
func CreateRecipientList(list *models.RecipientList) error {
	_, err := GetCollection("recipient_lists").InsertOne(context.Background(), list)
	return err
}

func FindRecipientList(id string) (models.RecipientList, error) {
	var list models.RecipientList
	err := GetCollection("recipient_lists").FindOne(context.Background(), bson.M{"_id": id}).Decode(&list)
	return list, err
}

func GetRecipientLists() ([]models.RecipientList, error) {
	cursor, err := GetCollection("recipient_lists").Find(context.Background(), map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	var lists []models.RecipientList
	if err = cursor.All(context.Background(), &lists); err != nil {
		return nil, err
	}
	return lists, nil
}

func UpdateRecipientList(listID string, list *models.RecipientList) error {
	_, err := GetCollection("recipient_lists").UpdateOne(
		context.Background(),
		bson.M{"_id": listID},
		bson.M{"$set": list},
	)
	return err
}

func DeleteRecipientList(listID string) error {
	_, err := GetCollection("recipient_lists").DeleteOne(
		context.Background(),
		bson.M{"_id": listID},
	)
	return err
}