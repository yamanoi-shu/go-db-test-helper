#!/bin/bash

SQL=$1
DB_NAME=$2

docker-compose exec -T test-db ${SQL} ${DB_NAME}
