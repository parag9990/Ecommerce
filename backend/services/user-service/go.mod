module github.com/parag/ecommerce/backend/services/user-service

go 1.24

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/go-sql-driver/mysql v1.9.3
	github.com/parag/ecommerce/backend/shared/gen/go v0.0.0
	google.golang.org/grpc v1.72.2
	google.golang.org/protobuf v1.36.6
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250428153025-10db94c68c34 // indirect
)

replace github.com/parag/ecommerce/backend/shared/gen/go => ../../shared/gen/go
