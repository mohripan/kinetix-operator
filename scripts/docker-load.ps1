$ErrorActionPreference = "Stop"

$cluster = $env:KINETIX_CLUSTER
if (-not $cluster) {
    $cluster = "kinetix"
}

$image = $env:KINETIX_IMAGE
if (-not $image) {
    $image = "kinetix/demo:dev"
}

kind load docker-image $image --name $cluster
