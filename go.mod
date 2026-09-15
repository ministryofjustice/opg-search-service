module github.com/ministryofjustice/opg-search-service

go 1.25.9

require (
	github.com/aws/aws-sdk-go-v2 v1.46.0
	github.com/aws/aws-sdk-go-v2/config v1.33.3
	github.com/aws/aws-sdk-go-v2/credentials v1.20.3
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.48.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.49.0
	github.com/aws/aws-secretsmanager-caching-go/v2 v2.3.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/gorilla/mux v1.8.1
	github.com/jackc/pgx/v5 v5.11.0
	github.com/opensearch-project/opensearch-go/v4 v4.7.3
	github.com/sirupsen/logrus v1.10.2
	github.com/stretchr/testify v1.12.1
)

require (
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.19.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.5.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.8.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.5.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.19 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.14.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.9.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.37.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.42.0 // indirect
	github.com/aws/smithy-go v1.28.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/text v0.31.0 // indirect
)
