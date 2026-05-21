# Scalable High-Performance E-Commerce Platform - Documentation Index

This repository is starting as a production blueprint for a microservices-based e-commerce platform using Golang, React, TypeScript, MongoDB, MySQL, Redis, Typesense, Kafka/RabbitMQ, Docker, and Kubernetes.

Language style: simple Hinglish with professional technical clarity.

## Step Mapping

| Step | Requirement | File |
|---:|---|---|
| 1 | Break project into micro tasks | [01-micro-tasks.md](./01-micro-tasks.md) |
| 2 | System architecture | [02-system-architecture.md](./02-system-architecture.md) |
| 3 | Folder structure | [03-folder-structure.md](./03-folder-structure.md) |
| 4 | Microservice design | [04-microservice-design.md](./04-microservice-design.md) |
| 5 | Master API JSON | [../api/master-api.json](../api/master-api.json) |
| 6 | Database design | [05-database-design.md](./05-database-design.md), [../database/draw.sql](../database/draw.sql), [../database/mongodb-schema-design.md](../database/mongodb-schema-design.md) |
| 7 | Auth and security | [06-auth-security.md](./06-auth-security.md) |
| 8 | Payment system | [07-payment-system.md](./07-payment-system.md) |
| 9 | Session management system | [08-session-management-system.md](./08-session-management-system.md) |
| 10 | CMS seller panel | [09-cms-superadmin.md](./09-cms-superadmin.md) |
| 11 | Frontend implementation | [10-frontend-implementation.md](./10-frontend-implementation.md) |
| 12 | DevOps | [11-devops-external-services.md](./11-devops-external-services.md) |
| 13 | External services | [11-devops-external-services.md](./11-devops-external-services.md) |
| 14 | Logging and monitoring | [12-logging-monitoring-scalability.md](./12-logging-monitoring-scalability.md) |
| 15 | Superadmin system | [09-cms-superadmin.md](./09-cms-superadmin.md) |
| 16 | Scalability design | [12-logging-monitoring-scalability.md](./12-logging-monitoring-scalability.md) |
| 17 | Documentation | [13-developer-guide.md](./13-developer-guide.md) |
| 18 | Git workflow and repository management | [14-git-workflow-repository-management.md](./14-git-workflow-repository-management.md) |

## Build Order Recommendation

1. Platform foundation: repo standards, Docker Compose, proto contracts, shared Go libraries.
2. Auth, User, Session, API Gateway.
3. Product, Search, CMS, Cart, Wishlist.
4. Order, Payment, Notification.
5. Recommendation, Analytics dashboards, Superadmin.
6. Kubernetes, CI/CD, monitoring, load testing, production hardening.
