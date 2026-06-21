# Local Ports

| Component | Container Port | Host Port | Protocol | Purpose | Env variable |
|---|---:|---:|---|---|---|
| API Gateway | 8080 | 8080 | HTTP | Public REST/liveness | `HTTP_ADDR` |
| Gateway gRPC-Web | 9090 | 8099 | HTTP/2 | Browser RPC facade | `GRPC_ADDR` |
| Gateway metrics | 9091 | 19090 | HTTP | Prometheus | `METRICS_ADDR` |
| Auth | 8081 | 8081 | HTTP | Auth/JWKS | `AUTH_HTTP_ADDR` |
| User | 50052 | 50052 | gRPC | User profiles | `USER_SERVICE_GRPC_ADDRESS` |
| Product | 8082 | 8082 | HTTP | Product REST/internal API and health | `PRODUCT_HTTP_ADDR` |
| Product | 9093 | 9092 | gRPC | Product catalog and inventory RPCs | `PRODUCT_GRPC_ADDR` |
| Cart | 8083 | 8083 | HTTP | Cart API | `CART_HTTP_ADDR` |
| Wishlist | 8084 | 8084 | HTTP | Wishlist API | `WISHLIST_HTTP_ADDR` |
| Search | 8085 | 8085 | HTTP | Search API | `SEARCH_HTTP_ADDR` |
| Session shell | 8086 | 8086 | HTTP | Local health only | `SESSION_HTTP_ADDR` |
| CMS HTTP / gRPC | 8087 / 9098 | 8087 / 50057 | HTTP / gRPC | Seller CMS | `CMS_HTTP_ADDR`, `CMS_GRPC_ADDR` |
| Recommendation HTTP / gRPC | 8088 / 9088 | 8089 / 50058 | HTTP / gRPC | Recommendations | `RECOMMENDATION_*_ADDR` |
| Order shell | 8090 | 8090 | HTTP | Local health only | `ORDER_LOCAL_HTTP_ADDR` |
| Payment | 8080 | 8091 | HTTP | Payments | `PAYMENT_HTTP_ADDR` |
| Notification HTTP / gRPC | 8081 / 9090 | 8092 / 50060 | HTTP / gRPC | Metrics/webhooks/RPC | `NOTIFICATION_*_ADDRESS` |
| Superadmin | 8088 | 8093 | HTTP | Admin API | `HTTP_ADDR` |
| User / Seller / Analytics / Admin UI | 8080 | 3000 / 3001 / 3002 / 3003 | HTTP | Frontends | build args |
| MySQL / MongoDB / Redis | standard | 3306 / 27017 / 6379 | TCP | Data stores | DSN/URI/address |
| RabbitMQ / UI | 5672 / 15672 | same | AMQP / HTTP | Events | RabbitMQ URL |
| Kafka | 29092 | 9092 | Kafka | Events | broker list |
| Typesense | 8108 | 8108 | HTTP | Search engine | `TYPESENSE_*` |
| Mailpit SMTP / UI | 1025 / 8025 | same | SMTP / HTTP | Local mail | SMTP vars |
| Jaeger UI / OTLP | 16686 / 4317 | same | HTTP / gRPC | Traces | OTLP endpoint |
| Prometheus | 9090 | 9095 | HTTP | Metrics UI | Compose mapping |

To fix a conflict, change the host port only, for example `"13306:3306"`, then use `localhost:13306` from host-run apps.
