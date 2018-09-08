#!/bin/bash
echo "=========>"

echo "stop docker"
docker-compose down

echo "clean docker"
docker stop $(docker ps -a -q)
docker rm $(docker ps -a -q)

echo "stop success"