// backend/db/mongo.go
package db

import (
	"context"
	"email-sender/backend/config"
	"email-sender/backend/models"
	"log"

	"go.mongodb.org/mongo-driver/bson" // Import bson for filtering
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

func CreateUser(user *models.User) error {
	_, err := GetCollection("users").InsertOne(context.Background(), user)
	return err
}

func FindUser(email string) (models.User, error) {
	var user models.User
	err := GetCollection("users").FindOne(context.Background(), map[string]string{"email": email}).Decode(&user)
	return user, err
}

func SaveEmailJob(job *models.EmailJob) error {
	_, err := GetCollection("email_jobs").InsertOne(context.Background(), job)
	return err
}

func SaveIPPair(pair *models.IPPair) error {
	_, err := GetCollection("ip_pairs").InsertOne(context.Background(), pair)
	return err
}

func GetIPPairs() ([]models.IPPair, error) {
	cursor, err := GetCollection("ip_pairs").Find(context.Background(), map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	var pairs []models.IPPair
	if err = cursor.All(context.Background(), &pairs); err != nil {
		return nil, err
	}
	return pairs, nil
}

func UpdateIPPair(subdomain, ip string, pair *models.IPPair) error {
	_, err := GetCollection("ip_pairs").UpdateOne(
		context.Background(),
		map[string]string{"subdomain": subdomain, "ip": ip},
		map[string]interface{}{"$set": pair},
	)
	return err
}

func DeleteIPPair(subdomain, ip string) error {
	_, err := GetCollection("ip_pairs").DeleteOne(
		context.Background(),
		map[string]string{"subdomain": subdomain, "ip": ip},
	)
	return err
}

// --- Functions for Campaign management ---

func CreateCampaign(campaign *models.Campaign) error {
	_, err := GetCollection("campaigns").InsertOne(context.Background(), campaign)
	return err
}

func FindCampaign(id string) (models.Campaign, error) {
	var campaign models.Campaign
	err := GetCollection("campaigns").FindOne(context.Background(), bson.M{"_id": id}).Decode(&campaign)
	return campaign, err
}

// --- Functions for RecipientList management ---

func CreateRecipientList(list *models.RecipientList) error {
	_, err := GetCollection("recipient_lists").InsertOne(context.Background(), list)
	return err
}

func FindRecipientList(id string) (models.RecipientList, error) {
	var list models.RecipientList
	err := GetCollection("recipient_lists").FindOne(context.Background(), bson.M{"_id": id}).Decode(&list)
	return list, err
}