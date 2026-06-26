# Kubernetes and Tilt Local Runbook

## Current support

Tilt manages the default runnable Docker Compose stack; the incomplete Session Analytics profile is omitted. Kubernetes currently has only a namespace/config/example-secret foundation. Full K8s application manifests would be misleading until the audited Product, Order, Session, and gateway transport blockers are fixed.

## Tilt with the working Compose stack

```bash
tilt up
tilt down
```

## Kubernetes foundation

Prerequisites: Docker, `kubectl`, and kind/minikube.

```bash
minikube start
kubectl get nodes
kubectl apply --dry-run=client -k infra/k8s/overlays/dev
kubectl apply -k infra/k8s/overlays/dev
kubectl get all -n ecommerce-local
kubectl delete namespace ecommerce-local
```

Future full setup must add Deployments, Services, PVCs, probes, and resource limits only after every application has a real runnable server contract. Tabhi `tilt up` ko Kubernetes image build/port-forward flow par switch karna safe hoga.
