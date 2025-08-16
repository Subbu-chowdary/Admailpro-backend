package smtp

import (
	"email-sender/backend/config"
	"email-sender/backend/metrics"
	"email-sender/backend/models"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
)

var sesClient *ses.SES

func init() {
	cfg := config.GetConfig()
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.AWSCredentials.Region),
		Credentials: credentials.NewStaticCredentials(cfg.AWSCredentials.AccessKey, cfg.AWSCredentials.SecretKey, ""),
	})
	if err != nil {
		log.Fatalf("Failed to create SES session: %v", err)
	}
	sesClient = ses.New(sess)
}

func SendEmail(job models.EmailJob) error {
	input := &ses.SendEmailInput{
		Destination: &ses.Destination{ToAddresses: []*string{aws.String(job.Request.Recipient)}},
		Message: &ses.Message{
			Body: &ses.Body{Html: &ses.Content{Charset: aws.String("UTF-8"), Data: aws.String(job.Request.HTML)}},
			Subject: &ses.Content{Charset: aws.String("UTF-8"), Data: aws.String(job.Request.Subject)},
		},
		Source: aws.String("sender@" + job.Subdomain),
	}
	_, err := sesClient.SendEmail(input)
	if err == nil {
		metrics.IncrementSentEmails()
	}
	return err
}