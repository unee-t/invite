#!/bin/bash
test -f "$1" || exit
QUEUE=https://sqs.ap-southeast-1.amazonaws.com/182387550209/invites
echo Attempting to put $1 onto $QUEUE
aws --profile ins-dev sqs send-message-batch \
	--queue-url $QUEUE \
	--entries file://$1
