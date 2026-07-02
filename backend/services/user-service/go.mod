module github.com/parag/ecommerce/backend/services/user-service

go 1.26.3

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/go-sql-driver/mysql v1.10.0
	github.com/parag/ecommerce/backend/shared/gen/go v0.0.0
	github.com/parag/ecommerce/backend/shared/platform v0.0.0
	github.com/parag/ecommerce/backend/shared/validation v0.0.0
	github.com/prometheus/client_golang v1.23.2
	github.com/rabbitmq/amqp091-go v1.11.0
	github.com/segmentio/kafka-go v0.4.51
	google.golang.org/grpc v1.81.1
	google.golang.org/protobuf v1.36.11
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nyaruka/phonenumbers v1.8.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260401024825-9d38bb4040a9 // indirect
)

replace github.com/parag/ecommerce/backend/shared/gen/go => ../../shared/gen/go

replace github.com/parag/ecommerce/backend/shared/platform => ../../shared/platform

replace github.com/parag/ecommerce/backend/shared/validation => ../../shared/validation
