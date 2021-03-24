#! /usr/bin/env bash

# Create ES domain
awslocal es create-elasticsearch-domain --domain-name opg --elasticsearch-version 7.9
