$ErrorActionPreference = "Stop"

$namespace = "kinetix"

kubectl apply -f deploy/local/namespace.yaml
kubectl apply -f deploy/local/prometheus.yaml
kubectl apply -f deploy/local/grafana.yaml
kubectl apply -f deploy/local/demo.yaml
kubectl -n $namespace rollout restart deployment/kinetix-producer deployment/kinetix-worker deployment/kinetix-sink

kubectl -n $namespace rollout status deployment/prometheus --timeout=120s
kubectl -n $namespace rollout status deployment/grafana --timeout=120s
kubectl -n $namespace rollout status deployment/kinetix-producer --timeout=180s
kubectl -n $namespace rollout status deployment/kinetix-worker --timeout=180s
kubectl -n $namespace rollout status deployment/kinetix-sink --timeout=180s

Write-Host "Prometheus: http://localhost:30000"
Write-Host "Grafana:    http://localhost:30001 (admin/admin)"
