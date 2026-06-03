module github.com/parag/ecommerce/backend/services/user-service

go 1.24

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/go-sql-driver/mysql v1.9.3
	github.com/parag/ecommerce/backend/shared/gen/go v0.0.0
	github.com/parag/ecommerce/backend/shared/validation v0.0.0
	github.com/rabbitmq/amqp091-go v1.11.0
	github.com/segmentio/kafka-go v0.4.51
	google.golang.org/grpc v1.72.2
	google.golang.org/protobuf v1.36.11
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/nyaruka/phonenumbers v1.8.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250428153025-10db94c68c34 // indirect
)

replace github.com/parag/ecommerce/backend/shared/gen/go => ../../shared/gen/go

replace github.com/parag/ecommerce/backend/shared/validation => ../../shared/validation
