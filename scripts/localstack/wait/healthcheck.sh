#!/usr/bin/env bash

# Elasticsearch
curl --silent http://localhost:4566/_localstack/health | grep "\"opensearch\": \"running\"" || exit 1
curl --write-out %{http_code} --silent --output /dev/null http://search-service.eu-west-1.opensearch.localhost.localstack.cloud:4566/_cluster/health | grep 200 || exit 1
awslocal opensearch list-domain-names | grep '"search-service"' || exit 1

curl -XPUT -H "Content-Type: application/json" http://search-service.eu-west-1.opensearch.localhost.localstack.cloud:4566/_cluster/settings -d '{ "transient": { "cluster.routing.allocation.disk.threshold_enabled": false } }'
echo "Disk thresholds disabled"
