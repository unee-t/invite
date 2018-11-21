package main

import (
	"context"
	"encoding/json"

	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/unee-t/invite"

	"github.com/apex/log"
	"github.com/aws/aws-lambda-go/lambda"
)

type SQStrigger struct {
	Records []struct {
		MessageID     string `json:"messageId"`
		ReceiptHandle string `json:"receiptHandle"`
		Body          string `json:"body"`
		Attributes    struct {
			ApproximateReceiveCount          string `json:"ApproximateReceiveCount"`
			SentTimestamp                    string `json:"SentTimestamp"`
			SenderID                         string `json:"SenderId"`
			ApproximateFirstReceiveTimestamp string `json:"ApproximateFirstReceiveTimestamp"`
		} `json:"attributes"`
		MessageAttributes struct {
		} `json:"messageAttributes"`
		Md5OfBody      string `json:"md5OfBody"`
		EventSource    string `json:"eventSource"`
		EventSourceARN string `json:"eventSourceARN"`
		AwsRegion      string `json:"awsRegion"`
	} `json:"Records"`
}

func main() {
	log.SetHandler(jsonhandler.Default)
	lambda.Start(handler)
}

func handler(ctx context.Context, evt SQStrigger) (string, error) {

	h, err := invite.New()
	if err != nil {
		return "", err
	}

	defer h.DB.Close()

	log.Infof("Number of records in trigger: %d", len(evt.Records))

	for i, v := range evt.Records {

		var ivt invite.Invite
		err := json.Unmarshal([]byte(v.Body), &ivt)

		log.Infof("Processing invite %d, %s", i, ivt.ID)
		err = h.ProcessInvite(ivt)
		if err != nil {
			return "", err
		}
	}

	return "", err
}
