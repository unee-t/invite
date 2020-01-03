#!/bin/bash
test -f "$1" || exit
# This is a hardcoded variable <-- should be moved
QUEUE=https://sqs.ap-southeast-1.amazonaws.com/812644853088/invites
#QUEUE=https://sqs.ap-southeast-1.amazonaws.com/915001051872/invites
# END This is a hardcoded variable
echo Attempting to put $1 onto $QUEUE
# This is a hardcoded variable <-- should be moved
aws --profile uneet-dev sqs send-message-batch \
# END This is a hardcoded variable
	--queue-url $QUEUE \
	--entries file://$1
