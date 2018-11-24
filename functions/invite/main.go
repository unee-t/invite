package main

import (
	"context"
	"encoding/json"

	jsonhandler "github.com/apex/log/handlers/json"
	"github.com/unee-t/invite"

	"github.com/apex/log"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	log.SetHandler(jsonhandler.Default)
	lambda.Start(handler)
}

func handler(ctx context.Context, evt events.SQSEvent) (string, error) {

	h, err := invite.New()
	if err != nil {
		return "", err
	}

	defer h.DB.Close()

	log.Infof("Number of records in trigger: %d", len(evt.Records))

	for i, v := range evt.Records {

		var ivt invite.Invite
		err := json.Unmarshal([]byte(v.Body), &ivt)

		log.Infof("Processing invite %d, Message ID: %s, Invite ID: %s", i, v.MessageId, ivt.ID)
		err = h.ProcessInvite(ivt)
		if err != nil {
			return "", err
		}
	}

	return "", err
}
