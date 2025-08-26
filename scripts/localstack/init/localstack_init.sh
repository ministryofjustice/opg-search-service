#! /usr/bin/env bash

# Create ES domain
awslocal opensearch create-domain --domain-name search-service

# Set secrets in Secrets Manager
awslocal secretsmanager create-secret --name /local/jwt-key \
   --description "JWT secret for Go services authentication" \
   --secret-string "MyTestSecret"

awslocal secretsmanager create-secret --name /local/user-hash-salt \
   --description "Email salt for Go services authentication" \
   --secret-string "ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0"
