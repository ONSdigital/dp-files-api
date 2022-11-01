#!/bin/bash
set -x
export AWS_ACCESS_KEY_ID=test
export AWS_SECRET_ACCESS_KEY=test
export AWS_DEFAULT_REGION=eu-west-1
awslocal s3 mb s3://testing
awslocal s3api put-object --bucket testing --key index.html --body /root/index.html
set +x