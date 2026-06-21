# Kubernetes platform foundation

`base/` contains reusable namespace and service manifests. `overlays/dev/` applies local development image tags and replica counts.

```bash
kubectl apply --dry-run=client -k infra/k8s/base
kubectl apply --dry-run=client -k infra/k8s/overlays/dev
kubectl apply -k infra/k8s/overlays/dev
```

The checked-in Secret manifests contain placeholders only. Replace them through an approved secret-management workflow before deploying outside an isolated development cluster.
