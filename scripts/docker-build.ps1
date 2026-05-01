$ErrorActionPreference = "Stop"

$image = $env:KINETIX_IMAGE
if (-not $image) {
    $image = "kinetix/demo:dev"
}

docker build -t $image ./workers
