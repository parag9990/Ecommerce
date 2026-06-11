module github.com/example/ecommerce-platform/backend/services/order-service

go 1.26.3

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/example/ecommerce-platform/backend/shared/gen/go v0.0.0
	github.com/go-sql-driver/mysql v1.9.3
	github.com/segmentio/kafka-go v0.4.51
	google.golang.org/grpc v1.81.1
	google.golang.org/protobuf v1.36.11
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260226221140-a57be14db171 // indirect
)

replace github.com/example/ecommerce-platform/backend/shared/gen/go => ../../shared/gen/go
