module ecommerce/superadmin-service

go 1.26.3

require (
	github.com/go-sql-driver/mysql v1.10.0
	github.com/parag/ecommerce/backend/shared/platform v0.0.0
)

replace github.com/parag/ecommerce/backend/shared/platform => ../../shared/platform

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
)
