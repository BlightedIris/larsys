$ImageName = "larsys-client-daemon"
$ContainerName = "larsys-client"

Write-Host "Building Docker image: $ImageName"
docker build -t $ImageName -f ./dev/client/Dockerfile .

Write-Host "Stopping existing container if running..."
docker stop $ContainerName 2>$null

Write-Host "Starting container: $ContainerName"
docker run --name larsys-client --network larsys-network -it --rm -v larsys-logs:/var/log/larsys -v larsys-tokens:/etc/larsys/tokens $ImageName

Write-Host "Loading container and opening logs"
docker exec -it larsys-server /bin/bash