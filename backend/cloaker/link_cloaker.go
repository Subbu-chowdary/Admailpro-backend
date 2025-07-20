// package cloaker

// import (
// 	"context"
// 	"strings"
// 	"time"

// 	"email-sender/backend/config"
// 	"email-sender/backend/models"

// 	"github.com/go-redis/redis/v8"
// 	"github.com/google/uuid"
// 	"golang.org/x/net/html"
// )

// var rdb *redis.Client

// func init() {
// 	cfg := config.GetConfig()
// 	rdb = redis.NewClient(&redis.Options{
// 		Addr: cfg.RedisAddr,
// 	})
// }

// func CloakLinks(htmlContent string) (string, []models.LinkMapping, error) {
// 	doc, err := html.Parse(strings.NewReader(htmlContent))
// 	if err != nil {
// 		return "", nil, err
// 	}

// 	var links []models.LinkMapping

// 	var f func(*html.Node)
// 	f = func(n *html.Node) {
// 		if n.Type == html.ElementNode && n.Data == "a" {
// 			for i, attr := range n.Attr {
// 				if attr.Key == "href" {
// 					id := uuid.New().String()
// 					cloaked := "https://track.mail3.example.com/redirect?id=" + id
// 					links = append(links, models.LinkMapping{
// 						ID:          id,
// 						OriginalURL: attr.Val,
// 					})
// 					n.Attr[i].Val = cloaked
// 				}
// 			}
// 		}
// 		for c := n.FirstChild; c != nil; c = c.NextSibling {
// 			f(c)
// 		}
// 	}
// 	f(doc)

// 	var b strings.Builder
// 	html.Render(&b, doc)

// 	ctx := context.Background()
// 	for _, link := range links {
// 		_ = rdb.Set(ctx, link.ID, link.OriginalURL, time.Hour*24*30).Err()
// 	}

//		return b.String(), links, nil
//	}
//
// backend/cloaker/link_cloaker.go
package cloaker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"email-sender/backend/config"
	"email-sender/backend/models"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"golang.org/x/net/html"
)

var rdb *redis.Client

func init() {
	cfg := config.GetConfig()
	rdb = redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})
}

func CloakLinks(htmlContent string) (string, []models.LinkMapping, error) {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", nil, err
	}

	var links []models.LinkMapping

	// Get the configured tracking domain
	cfg := config.GetConfig()
	trackingDomain := cfg.TrackingDomain

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for i, attr := range n.Attr {
				if attr.Key == "href" {
					id := uuid.New().String()
					// Use the configurable tracking domain
					cloaked := fmt.Sprintf("https://%s/redirect?id=%s", trackingDomain, id)
					links = append(links, models.LinkMapping{
						ID:          id,
						OriginalURL: attr.Val,
					})
					n.Attr[i].Val = cloaked
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	var b strings.Builder
	html.Render(&b, doc)

	ctx := context.Background()
	for _, link := range links {
		// Store link mappings in Redis with a 30-day expiration
		_ = rdb.Set(ctx, link.ID, link.OriginalURL, time.Hour*24*30).Err()
	}

	return b.String(), links, nil
}
