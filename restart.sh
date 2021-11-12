#! /bin/bash

echo "Pulling image"
docker pull ghcr.io/sirishbandi/blockchain:main

echo "Killing all containers"
for i in `docker ps -a |grep blockchain |awk '{print $1}'`;do docker rm $i -f;done;

echo "Starting Init"
CTR_ID=`docker run --net mynet -d ghcr.io/sirishbandi/blockchain:main ./blockchain --init=true --address=172.18.0.2:8080`
docker logs $CTR_ID
echo "Container ID: $CTR_ID"

echo ""
echo "writing data.."
curl 172.18.0.2:8080/addblock -d "testing"

echo "Sleeping for 25sec"
sleep 25
echo "Starting 2nd server"
CTR_ID=`docker run --net mynet --ip 172.18.0.5 -d ghcr.io/sirishbandi/blockchain:main ./blockchain  --address=172.18.0.2:8080 --myaddress=172.18.0.5:8080`
docker logs $CTR_ID
echo "Container ID: $CTR_ID"

