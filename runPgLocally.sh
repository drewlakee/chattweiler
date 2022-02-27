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
  MOUNT_SCRIPT="$(pwd)/sql/initdb.sql"
  MOUNT_SCRIPT=$(echo ${MOUNT_SCRIPT////\\})
  if [[ "$MOUNT_SCRIPT" == "\c"* ]]; then
      MOUNT_SCRIPT=$(echo ${MOUNT_SCRIPT/\\c/C:})
  elif [[ "$MOUNT_SCRIPT" == "\a"* ]]; then
      MOUNT_SCRIPT=$(echo ${MOUNT_SCRIPT/\a/A:})
  elif [[ "$MOUNT_SCRIPT" == "\b"* ]]; then
      MOUNT_SCRIPT=$(echo ${MOUNT_SCRIPT/\b/B:})
  elif [[ "$MOUNT_SCRIPT" == "\d"* ]]; then
      MOUNT_SCRIPT=$(echo ${MOUNT_SCRIPT/\d/D:})
  elif [[ "$MOUNT_SCRIPT" == "\f"* ]]; then
      MOUNT_SCRIPT=$(echo ${MOUNT_SCRIPT/\f/F:})
  else
    echo "runPgLocally.sh: Ooops, I don't know the name of your disk. Please, add me here to script and try again!"
    return
  fi

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
