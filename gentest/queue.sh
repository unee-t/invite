#!/bin/bash
test -f "$1" || exit
QUEUE=https://sqs.ap-southeast-1.amazonaws.com/812644853088/invites
echo Attempting to put $1 onto $QUEUE
aws --profile uneet-dev sqs send-message-batch \
	--queue-url $QUEUE \
	--entries file://$1
