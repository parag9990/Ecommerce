module product-service

go 1.26.3

require (
	github.com/parag/ecommerce/backend/shared/gen/go v0.0.0
	github.com/parag/ecommerce/backend/shared/platform v0.0.0
	github.com/rabbitmq/amqp091-go v1.11.0
	github.com/segmentio/kafka-go v0.4.51
	go.mongodb.org/mongo-driver/v2 v2.6.0
	google.golang.org/grpc v1.81.1
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	golang.org/x/crypto v0.51.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sync v0.20.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260401024825-9d38bb4040a9 // indirect
)

replace github.com/parag/ecommerce/backend/shared/gen/go => ../../shared/gen/go

replace github.com/parag/ecommerce/backend/shared/platform => ../../shared/platform
