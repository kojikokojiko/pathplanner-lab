#!/bin/bash
echo "Creating SQS queue..."
awslocal sqs create-queue --queue-name run-jobs
echo "SQS queue created."
