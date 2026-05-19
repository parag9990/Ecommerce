# Step 3 - Folder Structure

This file defines detailed folder structures for backend services, frontend apps, CMS dashboard, and session analytics dashboard.

## Monorepo Root

```text
ecommerce-platform/
  README.md
  docs/
  api/
    master-api.json
  proto/
    ecommerce/
      auth/v1/
      user/v1/
      product/v1/
      cart/v1/
      wishlist/v1/
      order/v1/
      payment/v1/
      search/v1/
      cms/v1/
      session/v1/
      notification/v1/
      superadmin/v1/
  backend/
    services/
    shared/
    tools/
  frontend/
    user-app/
    seller-dashboard/
    session-analytics-dashboard/
    superadmin-panel/
    packages/
  database/
    draw.sql
    mongodb-schema-design.md
    migrations/
  infra/
    docker/
    compose/
    k8s/
    helm/
    ci/
    monitoring/
```

## Backend - Golang Microservices

Clean architecture pattern:

- `cmd`: service entrypoint.
- `internal/domain`: core business entities and domain rules.
- `internal/usecase`: application workflows.
- `internal/repository`: DB access implementation.
- `internal/transport`: gRPC/REST handlers.
- `internal/config`: service config loader.
- `internal/middleware`: service-specific middleware.
- `internal/events`: message producer/consumer.
- `migrations`: SQL migrations if service uses MySQL.
- `deploy`: Dockerfile and service-specific K8s snippets.

```text
backend/
  go.work
  shared/
    config/
      loader.go
      env.go
    logger/
      logger.go
      zap.go
    errors/
      app_error.go
      grpc_mapping.go
      http_mapping.go
    authctx/
      claims.go
      context.go
    middleware/
      request_id.go
      recovery.go
      rate_limit.go
      tracing.go
    grpcclient/
      dialer.go
      interceptors.go
    validation/
      validator.go
    events/
      publisher.go
      consumer.go
      envelope.go
    idempotency/
      store.go
    health/
      health.go
  services/
    api-gateway/
      cmd/server/main.go
      internal/
        config/config.go
        routes/routes.go
        handlers/
          auth_handler.go
          product_handler.go
          cart_handler.go
          order_handler.go
        middleware/
          jwt.go
          rbac.go
          rate_limit.go
        clients/
          auth_client.go
          user_client.go
          product_client.go
      deploy/
        Dockerfile
        k8s.yaml
    auth-service/
      cmd/server/main.go
      internal/
        domain/
          account.go
          credential.go
          otp.go
          token.go
          role.go
        usecase/
          signup.go
          login.go
          refresh.go
          otp_verify.go
          logout.go
        repository/
          mysql_account_repository.go
          mysql_token_repository.go
          redis_rate_repository.go
        transport/
          grpc/server.go
          grpc/auth_handler.go
        events/
          publisher.go
        config/config.go
      migrations/
        001_create_auth_tables.up.sql
        001_create_auth_tables.down.sql
      deploy/
        Dockerfile
        k8s.yaml
    user-service/
      cmd/server/main.go
      internal/
        domain/
          user.go
          address.go
          seller_profile.go
        usecase/
          create_user.go
          update_profile.go
          manage_address.go
          seller_kyc.go
        repository/
          mysql_user_repository.go
        transport/
          grpc/server.go
      migrations/
      deploy/
    product-service/
      cmd/server/main.go
      internal/
        domain/
          product.go
          variant.go
          category.go
          inventory.go
        usecase/
          create_product.go
          publish_product.go
          reserve_inventory.go
          list_products.go
        repository/
          mongo_product_repository.go
          mongo_inventory_repository.go
        transport/
          grpc/server.go
        events/
          product_events.go
      deploy/
    cart-service/
      cmd/server/main.go
      internal/
        domain/cart.go
        usecase/
          add_item.go
          remove_item.go
          merge_cart.go
          price_cart.go
        repository/
          mongo_cart_repository.go
          redis_cart_cache.go
        transport/grpc/
      deploy/
    wishlist-service/
      cmd/server/main.go
      internal/
        domain/wishlist.go
        usecase/
        repository/mongo_wishlist_repository.go
        transport/grpc/
      deploy/
    order-service/
      cmd/server/main.go
      internal/
        domain/
          order.go
          order_item.go
          fulfillment.go
        usecase/
          create_order_from_cart.go
          cancel_order.go
          update_fulfillment.go
        repository/
          mysql_order_repository.go
          mysql_idempotency_repository.go
        transport/grpc/
        events/
      migrations/
      deploy/
    payment-service/
      cmd/server/main.go
      internal/
        domain/
          payment.go
          refund.go
          webhook_event.go
        usecase/
          create_payment_intent.go
          handle_webhook.go
          refund_payment.go
        provider/
          provider.go
          stripe_like.go
          razorpay_like.go
        repository/mysql_payment_repository.go
        transport/
          grpc/
          http/webhook_handler.go
      migrations/
      deploy/
    search-service/
      cmd/server/main.go
      internal/
        schema/typesense_schema.go
        indexer/product_indexer.go
        usecase/search_products.go
        repository/typesense_repository.go
        transport/grpc/
        events/product_consumer.go
      deploy/
    cms-service/
      cmd/server/main.go
      internal/
        domain/
          coupon.go
          campaign.go
          seller_setting.go
        usecase/
          validate_coupon.go
          manage_campaign.go
          seller_analytics.go
        repository/mysql_cms_repository.go
        transport/grpc/
      migrations/
      deploy/
    session-service/
      cmd/server/main.go
      internal/
        domain/
          session.go
          event.go
          journey.go
          heatmap.go
        ingest/
          event_ingestor.go
          user_agent_parser.go
        aggregation/
          funnel.go
          active_users.go
          heatmap.go
        repository/
          mongo_session_repository.go
          redis_active_session_repository.go
        transport/
          grpc/
          http/ingest_handler.go
      deploy/
    recommendation-service/
      cmd/server/main.go
      internal/
        domain/recommendation.go
        usecase/
          trending.go
          personalized.go
        repository/
          mongo_feature_repository.go
          redis_recommendation_cache.go
        events/interaction_consumer.go
        transport/grpc/
      deploy/
    notification-service/
      cmd/server/main.go
      internal/
        domain/
          template.go
          delivery.go
        usecase/
          send_otp.go
          send_order_update.go
        provider/
          email_provider.go
          sms_provider.go
        repository/mongo_notification_repository.go
        events/consumer.go
        transport/grpc/
      deploy/
    superadmin-service/
      cmd/server/main.go
      internal/
        domain/
          admin_user.go
          audit_log.go
          platform_setting.go
        usecase/
          manage_user.go
          manage_seller.go
          manage_payment.go
          audit_log.go
        repository/mysql_admin_repository.go
        transport/grpc/
      migrations/
      deploy/
  tools/
    proto-gen/
    migration-runner/
    load-test/
```

## Protobuf and gRPC Setup

```text
proto/
  buf.yaml
  buf.gen.yaml
  ecommerce/
    common/v1/
      pagination.proto
      money.proto
      errors.proto
      auth_context.proto
    auth/v1/auth.proto
    user/v1/user.proto
    product/v1/product.proto
    cart/v1/cart.proto
    wishlist/v1/wishlist.proto
    order/v1/order.proto
    payment/v1/payment.proto
    search/v1/search.proto
    cms/v1/cms.proto
    session/v1/session.proto
    notification/v1/notification.proto
    superadmin/v1/superadmin.proto
```

Generated code:

```text
backend/shared/gen/go/ecommerce/...
frontend/packages/proto-client/src/gen/...
```

## Frontend Shared Structure

Frontend monorepo packages:

```text
frontend/
  package.json
  pnpm-workspace.yaml
  tsconfig.base.json
  packages/
    ui/
      src/
        button.tsx
        input.tsx
        table.tsx
        modal.tsx
        toast.tsx
    api-client/
      src/
        http.ts
        auth.ts
        product.ts
        cart.ts
        order.ts
    proto-client/
      src/
        grpc-web.ts
        gen/
    auth/
      src/
        auth-store.ts
        protected-route.tsx
        role-guard.tsx
    config/
      src/
        env.ts
```

## User App - React + TypeScript

```text
frontend/user-app/
  index.html
  vite.config.ts
  tailwind.config.ts
  src/
    main.tsx
    app.tsx
    routes/
      index.tsx
      protected-routes.tsx
    app-shell/
      header.tsx
      mobile-nav.tsx
      search-box.tsx
      cart-badge.tsx
    features/
      auth/
        pages/
          login-page.tsx
          signup-page.tsx
          otp-page.tsx
        components/
        hooks/
          use-login.ts
          use-otp.ts
      product/
        pages/
          home-page.tsx
          category-page.tsx
          product-detail-page.tsx
        components/
          product-card.tsx
          product-grid.tsx
          filter-panel.tsx
        api/
          product.queries.ts
      search/
        pages/search-page.tsx
        components/facet-list.tsx
      cart/
        pages/cart-page.tsx
        components/cart-item-row.tsx
        api/cart.mutations.ts
      checkout/
        pages/
          checkout-page.tsx
          payment-result-page.tsx
        components/
          address-step.tsx
          payment-step.tsx
      wishlist/
      profile/
      orders/
    stores/
      ui-store.ts
      auth-store.ts
    lib/
      query-client.ts
      grpc-client.ts
      analytics.ts
    styles/
      globals.css
```

## CMS Dashboard - Seller Panel

```text
frontend/seller-dashboard/
  vite.config.ts
  tailwind.config.ts
  src/
    main.tsx
    app.tsx
    layout/
      dashboard-layout.tsx
      sidebar.tsx
      topbar.tsx
    routes/
      seller-routes.tsx
    features/
      overview/
        pages/overview-page.tsx
        components/revenue-card.tsx
      products/
        pages/
          product-list-page.tsx
          product-editor-page.tsx
        components/
          variant-editor.tsx
          image-uploader.tsx
          category-selector.tsx
      orders/
        pages/order-list-page.tsx
        pages/order-detail-page.tsx
      offers/
        pages/coupon-list-page.tsx
        pages/coupon-editor-page.tsx
      analytics/
        pages/revenue-analytics-page.tsx
      team/
        pages/team-members-page.tsx
      audit/
        pages/activity-log-page.tsx
    api/
      cms-api.ts
      seller-product-api.ts
    stores/
      seller-store.ts
```

## Session Analytics Dashboard

```text
frontend/session-analytics-dashboard/
  vite.config.ts
  tailwind.config.ts
  src/
    main.tsx
    app.tsx
    layout/
      analytics-layout.tsx
      filters-bar.tsx
    features/
      live/
        pages/live-sessions-page.tsx
        components/active-user-map.tsx
      journey/
        pages/journey-explorer-page.tsx
        components/session-timeline.tsx
      funnels/
        pages/funnel-analysis-page.tsx
        components/funnel-chart.tsx
      heatmaps/
        pages/heatmap-page.tsx
        components/heatmap-canvas.tsx
      cohorts/
        pages/cohort-retention-page.tsx
      privacy/
        pages/privacy-controls-page.tsx
    api/
      session-api.ts
    lib/
      chart-theme.ts
```

## Superadmin Panel

```text
frontend/superadmin-panel/
  vite.config.ts
  tailwind.config.ts
  src/
    main.tsx
    app.tsx
    layout/
      admin-layout.tsx
      admin-sidebar.tsx
    features/
      users/
        pages/user-list-page.tsx
        pages/user-detail-page.tsx
      sellers/
        pages/seller-list-page.tsx
        pages/seller-review-page.tsx
      orders/
        pages/order-operations-page.tsx
      payments/
        pages/payment-operations-page.tsx
        pages/refund-review-page.tsx
      sessions/
        pages/session-oversight-page.tsx
      settings/
        pages/platform-settings-page.tsx
      audit/
        pages/admin-audit-log-page.tsx
    api/
      superadmin-api.ts
```

## Config Management

Each Go service should load config in this order:

1. Defaults from code.
2. Local `.env` for development.
3. Kubernetes ConfigMap.
4. Kubernetes Secret or external secret manager for secrets.

Example service environment:

```text
SERVICE_NAME=product-service
SERVICE_PORT=8080
GRPC_PORT=9090
MONGO_URI=mongodb://product-mongo:27017/product_db
REDIS_ADDR=redis:6379
KAFKA_BROKERS=kafka:9092
LOG_LEVEL=info
TRACE_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4317
```

## Middleware Stack

Gateway HTTP middleware order:

1. Request ID
2. Real IP / forwarded headers
3. Panic recovery
4. Structured logging
5. Metrics
6. Tracing
7. CORS
8. Rate limiting
9. Authentication
10. RBAC
11. Request validation

Service gRPC interceptor order:

1. Request ID propagation
2. Deadline enforcement
3. Panic recovery
4. Structured logging
5. Metrics
6. Tracing
7. Auth context extraction
8. Validation

