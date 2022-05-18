#!/bin/bash

DB_NAME=$1

SQL="DROP DATABASE IF EXISTS ${DB_NAME}; CREATE DATABASE IF NOT EXISTS ${DB_NAME}"

docker-compose exec -T test-db ${SQL}
