$ErrorActionPreference = "Stop"

$namespace = "kinetix"
$strimziVersion = "0.45.0"
$daprVersion = "1.15.8"

kubectl apply -f deploy/local/namespace.yaml

helm repo add strimzi https://strimzi.io/charts/
helm repo add dapr https://dapr.github.io/helm-charts/
helm repo update

helm upgrade --install strimzi-kafka-operator strimzi/strimzi-kafka-operator `
    --namespace strimzi-system `
    --create-namespace `
    --version $strimziVersion `
    --set watchAnyNamespace=true

helm upgrade --install dapr dapr/dapr `
    --namespace dapr-system `
    --create-namespace `
    --version $daprVersion `
    --wait

kubectl -n strimzi-system rollout status deployment/strimzi-cluster-operator --timeout=300s
kubectl apply -f deploy/local/redis.yaml
kubectl -n $namespace rollout status deployment/redis --timeout=120s

kubectl apply -f deploy/local/dapr-components.yaml
kubectl apply -f deploy/local/strimzi/kafka.yaml
kubectl -n $namespace wait kafka.kafka.strimzi.io/kinetix-kafka --for=condition=Ready --timeout=300s
kubectl -n $namespace wait kafkatopic.kafka.strimzi.io/kinetix-input --for=condition=Ready --timeout=120s
kubectl -n $namespace wait kafkatopic.kafka.strimzi.io/kinetix-output --for=condition=Ready --timeout=120s
