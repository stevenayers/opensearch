package queue

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/go-kit/kit/log/level"
	"github.com/stevenayers/clamber/pkg/config"
	"github.com/stevenayers/clamber/pkg/logging"
	"github.com/stevenayers/clamber/pkg/page"
)

type Queue struct {
	ReceiveChan chan *sqs.Message
	Svc         *sqs.SQS
}

func NewQueue() (queue *Queue) {

	myCustomResolver := func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		if service == endpoints.SqsServiceID {
			return endpoints.ResolvedEndpoint{
				URL:           "http://localhost:4100",
				SigningRegion: "faux-region-1",
			}, nil
		}

		return endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:           aws.String("us-west-2"),
		EndpointResolver: endpoints.ResolverFunc(myCustomResolver),
	}))
	queue = &Queue{
		ReceiveChan: make(chan *sqs.Message, config.AppConfig.Queue.MaxConcurrentReceivedMessages),
		Svc:         sqs.New(sess),
	}
	return
}

func (q *Queue) Poll() {
	for {
		output, err := q.Svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			AttributeNames: []*string{
				aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
			},
			MessageAttributeNames: []*string{
				aws.String(sqs.QueueAttributeNameAll),
			},
			QueueUrl:            &config.AppConfig.Queue.QueueURL,
			MaxNumberOfMessages: aws.Int64(config.AppConfig.Queue.MaxConcurrentReceivedMessages),
			WaitTimeSeconds:     aws.Int64(config.AppConfig.Queue.SQSWaitTimeSeconds),
		})

		if err != nil {
			_ = level.Error(logging.Logger).Log("msg", err.Error())
			return
		}

		for _, message := range output.Messages {
			q.ReceiveChan <- message

			_, err := q.Svc.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      &config.AppConfig.Queue.QueueURL,
				ReceiptHandle: message.ReceiptHandle,
			})

			if err != nil {
				_ = level.Error(logging.Logger).Log("msg", err.Error())
				return
			} else {
				_ = level.Info(logging.Logger).Log("msg", "Successfully deleted message", "messageId", *message.MessageId, "payload", *message.Body)
			}
		}
	}
}

func (q *Queue) Publish(p *page.Page) {
	sqsPage := page.ConvertPageToSQSPage(p)
	if p.Parent != nil {
		sqsPage.Parent = page.ConvertPageToSQSPage(p.Parent)
	}
	payload, err := json.Marshal(sqsPage)
	result, err := q.Svc.SendMessage(&sqs.SendMessageInput{
		DelaySeconds: aws.Int64(10),
		MessageBody:  aws.String(string(payload)),
		QueueUrl:     &config.AppConfig.Queue.QueueURL,
	})
	if err != nil {
		_ = level.Error(logging.Logger).Log("msg", err.Error())
		return
	}
	_ = level.Info(logging.Logger).Log("msg", "Successfully sent message", "messageId", *result.MessageId, "url", p.Url, "start_url", p.StartUrl)
}
