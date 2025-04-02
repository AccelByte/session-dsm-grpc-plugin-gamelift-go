#!/bin/bash

# These variables must be exported for `extend-helper-cli` to work
export AB_BASE_URL="YOUR_BASE_URL_HERE"
export AB_CLIENT_ID="YOUR_CLIENT_ID_HERE"
export AB_CLIENT_SECRET="YOUR_CLIENT_SECRET_HERE"
export REPO_URL="YOUR_REPO_URL_HERE"

# Replace these AccelByte variables with your own values
AB_NAMESPACE="YOUR_ACCELBYTE_NAMESPACE_HERE"
AB_EXTEND_APP_NAME="bytewars-session-dsm"
AB_EXTEND_APP_IMAGE_TAG="0.0.0"

# Replace these AWS variables with your own values
AWS_REGION="us-west-2"
AWS_ACCESS_KEY_ID="YOUR_AWS_ACCESS_KEY_ID_HERE"
AWS_SECRET_ACCESS_KEY="YOUR_AWS_SECRET_ACCESS_KEY_HERE"

# ---

# Create the Session DSM app
extend-helper-cli create-app \
	--namespace $AB_NAMESPACE --app $AB_EXTEND_APP_NAME \
	--scenario function-override --confirm --wait

# Configure environment secrets
extend-helper-cli update-var \
	--namespace $AB_NAMESPACE --app $AB_EXTEND_APP_NAME \
	--key AWS_REGION --value $AWS_REGION
extend-helper-cli update-var \
	--namespace $AB_NAMESPACE --app $AB_EXTEND_APP_NAME \
	--key AWS_ACCESS_KEY_ID --value $AWS_ACCESS_KEY_ID
extend-helper-cli update-var \
	--namespace $AB_NAMESPACE --app $AB_EXTEND_APP_NAME \
	--key AWS_SECRET_ACCESS_KEY--value $AWS_SECRET_ACCESS_KEY

# Authenticate with Docker, then build and upload the Session DSM image
extend-helper-cli image-upload \
	--namespace $AB_NAMESPACE --app $AB_EXTEND_APP_NAME \
	--login --image-tag $AB_EXTEND_APP_IMAGE_TAG

# Deploy the Extend App
extend-helper-cli deploy-app \
	--namespace $AB_NAMESPACE --app $AB_EXTEND_APP_NAME \
	--image-tag $AB_EXTEND_APP_IMAGE_TAG --wait