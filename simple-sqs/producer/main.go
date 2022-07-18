package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

const qname = "maksim-kremenev-test-1.fifo"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		log.Panic(err)
	}

	svc := sqs.New(sess)

	output, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(qname),
	})
	if err != nil {
		log.Panic(err)
	}

	for i := 0; i < 1; i++ {
		body := fmt.Sprintf("Current time is: %s", time.Now().Format(time.RFC3339))
		if _, err := svc.SendMessage(&sqs.SendMessageInput{
			MessageBody:            &body,
			QueueUrl:               output.QueueUrl,
			MessageGroupId:         aws.String("group-1"),
			MessageDeduplicationId: aws.String(time.Now().Format(time.RFC3339)),
		}); err != nil {
			log.Panic(err)
		}
		log.Printf("Send a message: %s", body)
	}

	log.Println("Done")
}
