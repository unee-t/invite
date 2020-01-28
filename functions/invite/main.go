package main

import (
	"context"
	"encoding/json"

	"github.com/apex/log"
	jsonhandler "github.com/apex/log/handlers/json"
// This is a hardcoded variable <-- should be moved
	"github.com/unee-t-ins/invite"
// END This is a hardcoded variable <-- should be moved
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	log.SetHandler(jsonhandler.Default)
	lambda.Start(handler)
}

func handler(ctx context.Context, evt events.SQSEvent) (string, error) {

	h, err := invite.New(ctx)
	if err != nil {
		return "", err
	}
	h.Log.Infof("Got database connection")

	defer h.DB.Close()

	h.Log.Infof("Number of records in trigger: %d", len(evt.Records))

	for _, v := range evt.Records {

		var ivt invite.Invite
		err := json.Unmarshal([]byte(v.Body), &ivt)

		log.WithFields(log.Fields{
			"invite": ivt,
			"msgid":  v.MessageId,
		}).Info("processing invite via lambda")

		err = h.ProcessInvite(ivt)
		if err != nil {
			return "", err
		}
	}

	return "", err
}
