$ImageName = "larsys-daemon"
$ContainerName = "larsys-server"

Write-Host "Stopping $ContainerName in case it is running."
docker container stop $ContainerName

Write-Host "Building Docker image: $ImageName"
docker build -t $ImageName -f ./dev/host/Dockerfile .

Write-Host "Stopping existing container if running..."
docker stop $ContainerName 2>$null

Write-Host "Starting container: $ContainerName"
docker run --name $ContainerName --network larsys-network -d -it --rm  -v larsys-tokens:/etc/larsys/tokens -v larsys-logs:/var/log/larsys $ImageName

Write-Host "Loading container and opening logs"
docker exec -it larsys-server /bin/bash -lc "tail -f /var/log/larsys/daemon.log & exec /bin/bash"