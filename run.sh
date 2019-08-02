#!/bin/bash

source aws-set-profile skycatch
source aws-set-role dev

docker run -it --rm \
-e 'RETURN_RESULT=true' \
-e 'TASK_ACTIVITY_ARN=arn:aws:states:us-west-2:533689966658:activity:demo-sfn-response-activity' \
-e "AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}" \
-e "AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}" \
-e "AWS_SESSION_TOKEN=${AWS_SESSION_TOKEN}" \
nodetask:latest