#!/bin/bash

aws dynamodb create-table \
    --cli-input-json file://dynamodb-schema.json \
    --billing-mode PAY_PER_REQUEST \
    --region us-east-1