module github.com/example/ecommerce-platform/backend/services/auth-service

go 1.26.3

require (
	github.com/go-sql-driver/mysql v1.10.0
	github.com/parag/ecommerce/backend/shared/gen/go v0.0.0
	golang.org/x/crypto v0.51.0
	google.golang.org/grpc v1.81.1
)

replace github.com/parag/ecommerce/backend/shared/gen/go => ../../shared/gen/go

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260226221140-a57be14db171 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)
