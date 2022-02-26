#!/bin/bash

if [[ "$OSTYPE" == "linux"* ]]; then
  MOUNT_SCRIPT="$(pwd)/sql/initdb.sql"
  docker run -p 5433:5432 \
  -v ./sql/initdb.sql:/docker-entrypoint-initdb.d/initdb.sql \
  --name postgres-chattweiler \
  -e POSTGRES_DB=chattweiler \
  -e POSTGRES_USER=${1:-postgres} \
  -e POSTGRES_PASSWORD=${2:-postgres} \
  -d postgres
elif [[ "$OSTYPE" == "win"* || "$OS" == "Windows"* ]]; then
  MOUNT_SCRIPT="C:\Users\\$USERNAME\go\src\chattweiler\sql\initdb.sql"
  docker run -p 5433:5432 \
  -v "$MOUNT_SCRIPT":/docker-entrypoint-initdb.d/initdb.sql \
  --name postgres-chattweiler \
  -e POSTGRES_DB=chattweiler \
  -e POSTGRES_USER=${1:-postgres} \
  -e POSTGRES_PASSWORD=${2:-postgres} \
  -d postgres
else
  echo "Unknown OS for that script scenario"
fi
