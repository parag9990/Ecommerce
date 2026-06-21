module github.com/example/ecommerce-platform/backend/services/order-service

go 1.26.3

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/example/ecommerce-platform/backend/shared/gen/go v0.0.0
	github.com/go-sql-driver/mysql v1.10.0
	github.com/segmentio/kafka-go v0.4.51
	google.golang.org/grpc v1.81.1
	google.golang.org/protobuf v1.36.11
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/xdg-go/scram v1.2.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260226221140-a57be14db171 // indirect
)

replace github.com/example/ecommerce-platform/backend/shared/gen/go => ../../shared/gen/go
