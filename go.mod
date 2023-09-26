module github.com/ministryofjustice/opg-search-service

go 1.16

require (
	github.com/aws/aws-sdk-go v1.44.279
	github.com/aws/aws-secretsmanager-caching-go v1.1.0
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/gorilla/mux v1.8.0
	github.com/jackc/pgx/v4 v4.18.1
	github.com/lib/pq v1.10.4 // indirect
	github.com/ministryofjustice/opg-go-common v0.0.0-20220816144329-763497f29f90
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.8.4
	go.opentelemetry.io/contrib/detectors/aws/ecs v1.19.0
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.44.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.44.0
	go.opentelemetry.io/contrib/propagators/aws v1.19.0
	go.opentelemetry.io/otel v1.18.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.18.0
	go.opentelemetry.io/otel/sdk v1.18.0
	google.golang.org/grpc v1.58.2
)
