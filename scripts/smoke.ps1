$ErrorActionPreference = "Stop"

$namespace = "kinetix"
$job = "kinetix-smoke"
$image = $env:KINETIX_IMAGE
if (-not $image) {
    $image = "kinetix/demo:dev"
}

kubectl -n $namespace delete job $job --ignore-not-found
kubectl -n $namespace create job $job --image=$image -- `
    /sink `
    -brokers kinetix-kafka-kafka-bootstrap.kinetix.svc.cluster.local:9092 `
    -topic kinetix-output `
    -group "kinetix-smoke-$([DateTimeOffset]::UtcNow.ToUnixTimeSeconds())" `
    -count 5 `
    -timeout 90s

kubectl -n $namespace wait --for=condition=complete "job/$job" --timeout=120s
kubectl -n $namespace logs "job/$job"
