$ErrorActionPreference = "Stop"

$cluster = $env:KINETIX_CLUSTER
if (-not $cluster) {
    $cluster = "kinetix"
}

kind delete cluster --name $cluster
