module github.com/example/ecommerce-platform/backend/services/search-service

go 1.26.3

require (
	github.com/parag/ecommerce/backend/shared/gen/go v0.0.0
	github.com/parag/ecommerce/backend/shared/platform v0.0.0
	github.com/prometheus/client_golang v1.23.2
	github.com/rabbitmq/amqp091-go v1.11.0
	github.com/typesense/typesense-go/v2 v2.0.0
	google.golang.org/grpc v1.81.1
	google.golang.org/protobuf v1.36.11
)

replace github.com/parag/ecommerce/backend/shared/gen/go => ../../shared/gen/go

replace github.com/parag/ecommerce/backend/shared/platform => ../../shared/platform

// typesense-go/v2 still requests the pre-split genproto module. Keep the
// parent module aligned with the workspace to avoid duplicate RPC packages.
replace google.golang.org/genproto => google.golang.org/genproto v0.0.0-20241113202542-65e8d215514f

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oapi-codegen/runtime v1.1.1 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/sony/gobreaker v0.5.0 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.44.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260401024825-9d38bb4040a9 // indirect
)
