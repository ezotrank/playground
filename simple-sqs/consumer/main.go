package main

import (
	"fmt"
	"log"

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

	for {
		msgResult, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl: output.QueueUrl,
		})
		if err != nil {
			log.Panic(err)
		}
		if len(msgResult.Messages) == 0 {
			continue
		}

		fmt.Println(msgResult.Messages)

		_, err = svc.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      output.QueueUrl,
			ReceiptHandle: msgResult.Messages[0].ReceiptHandle,
		})
	}
}
