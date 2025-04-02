#!/bin/bash

# These variables must be exported for `extend-helper-cli` to work
export AB_BASE_URL="YOUR_BASE_URL_HERE"
export AB_CLIENT_ID="YOUR_CLIENT_ID_HERE"
export AB_CLIENT_SECRET="YOUR_CLIENT_SECRET_HERE"
export REPO_URL="YOUR_REPO_URL_HERE"

# Replace these AccelByte variables with your own values
AB_NAMESPACE="YOUR_ACCELBYTE_NAMESPACE_HERE"
AB_EXTEND_APP_NAME="bytewars-session-dsm"

# Replace these AWS variables with your own values
AWS_LOCATION_OVERRIDE=""
AWS_ALIAS_ID_OVERRIDE=""
AWS_QUEUE_ARN_OVERRIDE=""

if [ -n "$AWS_LOCATION_OVERRIDE" ];
    extend-helper-cli update-var \
	    --namespace $AB_NAMESPACE --app $AB_EXTEND_APP_NAME \
	    --key AWS_LOCATION_OVERRIDE --value $AWS_LOCATION_OVERRIDE
fi

if [ -n "$AWS_ALIAS_ID_OVERRIDE" ];
    extend-helper-cli update-var \
        --namespace $AB_NAMESPACE --app $AB_EXTEND_APP_NAME \
        --key AWS_ALIAS_ID_OVERRIDE --value $AWS_ALIAS_ID_OVERRIDE
fi

if [ -n "$AWS_QUEUE_ARN_OVERRIDE" ];
    extend-helper-cli update-var \
        --namespace $AB_NAMESPACE --app $AB_EXTEND_APP_NAME \
        --key AWS_QUEUE_ARN_OVERRIDE --value $AWS_QUEUE_ARN_OVERRIDE
fi
