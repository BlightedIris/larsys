$SERVER="larsys-server"
$CLIENT="larsys-client"
$NETWORK="larsys-network"

Write-Host "Stopping $SERVER"
docker stop $SERVER

Write-Host "Stopping $CLIENT"
docker stop $CLIENT

Write-Host "Stopping $NETWORK"
docker network remove $NETWORK

Write-Host "Starting $NETWORK"
docker network create $NETWORK