#!/bin/bash
echo "========>"
echo "clean docker"
docker stop $(docker ps -a -q)
docker rm $(docker ps -a -q)

rm all.log

echo "start docker"
docker-compose up >> all.log 2>&1 &
echo "start success..."