# Local Kubernetes Foundation

This local overlay includes the namespace/config/secret foundation plus runnable API Gateway and Product Service workloads. Datastores remain external to this overlay; Docker Compose is still the easiest full-stack path.

```bash
kubectl apply -k deployments/k8s/local
kubectl get all -n ecommerce-local
kubectl get ingress -n ecommerce-local
kubectl delete namespace ecommerce-local
```

Never replace the example dummy Secret with real production credentials in Git.
