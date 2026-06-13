# Search Reindex Job

Apply the namespace, config, secret, and job manifests from this folder to run a full catalog reindex as a Kubernetes Job.

```bash
kubectl apply -f infra/k8s/jobs/namespace.yaml
kubectl apply -f infra/k8s/jobs/search-reindex/configmap.yaml
kubectl apply -f infra/k8s/jobs/search-reindex/job.yaml
```

Create `search-reindex-secret` from the real secret manager before applying the Job. `secret.example.yaml` documents the required keys only.
