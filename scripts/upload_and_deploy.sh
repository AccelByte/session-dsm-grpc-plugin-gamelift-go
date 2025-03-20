#!/bin/bash

# These variables must be exported for `extend-helper-cli` to work
export AB_BASE_URL="YOUR_BASE_URL_HERE"
export AB_CLIENT_ID="YOUR_CLIENT_ID_HERE"
export AB_CLIENT_SECRET="YOUR_CLIENT_SECRET_HERE"
export REPO_URL="YOUR_REPO_URL_HERE"

AB_NAMESPACE="YOUR_ACCELBYTE_NAMESPACE_HERE"
AB_EXTEND_APP_NAME="bytewars-session-dsm"
AB_EXTEND_APP_IMAGE_TAG="0.0.0"

# Authenticate with Docker, then build and upload the Session DSM image
extend-helper-cli image-upload \
	--namespace $AB_NAMESPACE --app $AB_EXTEND_APP_NAME \
	--login --image-tag $AB_EXTEND_APP_IMAGE_TAG

# Deploy the Extend App
extend-helper-cli deploy-app \
	--namespace $AB_NAMESPACE --app $AB_EXTEND_APP_NAME \
	--image-tag $AB_EXTEND_APP_IMAGE_TAG --wait
