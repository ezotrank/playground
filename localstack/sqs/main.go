package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const (
	Endpoint        = "http://localhost:4566"
	AccessKeyID     = "secret"
	SecretAccessKey = "secret"
	Region          = "ap-southeast-1"
)

const qname = "test-queue.fifo"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	sess, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(Endpoint),
		Region:   aws.String(Region),
	})
	if err != nil {
		log.Panic(err)
	}

	svc := sqs.New(sess)

	result, err := svc.CreateQueue(&sqs.CreateQueueInput{
		QueueName: aws.String(qname),
		Attributes: map[string]*string{
			"DelaySeconds":           aws.String("60"),
			"MessageRetentionPeriod": aws.String("86400"),
		},
	})
	if err != nil {
		log.Panic(err)
	}

	if _, err := svc.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String("AAAA Information about current NY Times fiction bestseller for week of 12/11/2016."),
		QueueUrl:    result.QueueUrl,
	}); err != nil {
		log.Panic(err)
	}

	msgResult, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		MessageAttributeNames: []*string{
			aws.String(sqs.QueueAttributeNameAll),
		},
		QueueUrl:            result.QueueUrl,
		MaxNumberOfMessages: aws.Int64(1),
		VisibilityTimeout:   aws.Int64(20),
	})
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(msgResult.Messages)
}
