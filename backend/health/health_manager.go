package health

import (
	"context"
	"email-sender/backend/config"
	"email-sender/backend/db"
	"email-sender/backend/models"
	"log"
	"math"
	"sync"
)

type HealthManager struct {
	pairs []models.IPPair
	mutex sync.Mutex
}

// Initialize the HealthManager
func NewHealthManager() *HealthManager {
	hm := &HealthManager{}
	cursor, err := db.GetCollection("ip_pairs").Find(context.Background(), map[string]interface{}{})
	if err != nil {
		log.Fatal("Failed to fetch ip_pairs:", err)
	}

	if err = cursor.All(context.Background(), &hm.pairs); err != nil {
		log.Fatal("Failed to decode ip_pairs:", err)
	}

	// Fallback to default config if DB is empty
	if len(hm.pairs) == 0 {
		hm.pairs = generateDefaultPairs()
		for _, pair := range hm.pairs {
			_ = db.SaveIPPair(&pair)
		}
	}

	return hm
}

// Generate default IP/subdomain pairs from env-config
func generateDefaultPairs() []models.IPPair {
	var pairs []models.IPPair
	for _, cfg := range config.GetConfig().SESConfigs {
		pairs = append(pairs, models.IPPair{
			Subdomain: cfg.Subdomain,
			IP:        cfg.IP,
			Health:    90,
			SentCount: 0,
		})
	}
	return pairs
}

// Update health score for an IP/subdomain pair
func (hm *HealthManager) UpdateHealth(subdomain, ip string, spamRate, bounceRate, complaintRate float64) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	for i, pair := range hm.pairs {
		if pair.Subdomain == subdomain && pair.IP == ip {
			// Health calculation formula
			health := 100 - (spamRate*45 + bounceRate*35 + complaintRate*20 + math.Sqrt(float64(pair.SentCount))*0.1)
			pair.Health = math.Max(0, math.Min(100, health))
			pair.SentCount++

			hm.pairs[i] = pair
			_ = db.UpdateIPPair(subdomain, ip, &pair)
			break
		}
	}
}

// Get healthiest subdomain+IP pair for sending
func (hm *HealthManager) GetHealthiestPair() models.IPPair {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	var bestPair models.IPPair
	maxHealth := float64(-1)

	for _, p := range hm.pairs {
		if p.Health > maxHealth {
			maxHealth = p.Health
			bestPair = p
		}
	}

	return bestPair
}
