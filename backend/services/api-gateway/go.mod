module github.com/parag/ecommerce/backend/services/api-gateway

go 1.24

require (
	github.com/parag/ecommerce/backend/shared/gen/go v0.0.0
	github.com/parag/ecommerce/backend/shared/validation v0.0.0
	google.golang.org/grpc v1.72.2
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/nyaruka/phonenumbers v1.8.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250428153025-10db94c68c34 // indirect
)

replace github.com/parag/ecommerce/backend/shared/gen/go => ../../shared/gen/go

replace github.com/parag/ecommerce/backend/shared/validation => ../../shared/validation
