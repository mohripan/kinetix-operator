$ErrorActionPreference = "Stop"

$cluster = $env:KINETIX_CLUSTER
if (-not $cluster) {
    $cluster = "kinetix"
}

$existing = kind get clusters | Where-Object { $_ -eq $cluster }
if ($existing) {
    Write-Host "kind cluster '$cluster' already exists."
    kubectl config use-context "kind-$cluster"
    exit 0
}

kind create cluster --name $cluster --config deploy/local/kind/cluster.yaml
kubectl config use-context "kind-$cluster"
