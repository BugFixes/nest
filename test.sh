#!/usr/bin/env bash

function createDatabase()
{
  echo "createDatabase"
  docker run \
    -d \
    -p 5432:5432 \
    -e POSTGRES_PASSWORD=tester \
    -e POSTGRES_USERNAME=postgres \
    -e POSTGRES_DB=postgres \
    --name tester_postgres \
    postgres:11.5
}

function injectStructure()
{
  echo "injectStructure"
  docker exec \
    -e PGPASSWORD=tester tester_postgres psql \
    -U postgres \
    -d postgres \
    -c "CREATE TABLE "public"."bug" ("id" uuid, "hash" text, "message" text, "agent_id" uuid, "level" int, "time_posted" timestamp, PRIMARY KEY ("id"));"
}

function dropStructure()
{
  echo "dropStructure"
  docker exec \
    -e PGPASSWORD=tester tester_postgres psql \
    -U postgres \
    -d postgres \
    -c "DROP TABLE "public"."bug";"
}

function createQueue()
{
  echo "createQueue"
  awslocal sqs create-queue \
    --queue-name tester
}

function testCode()
{
  echo "testCode"
  DB_DATABASE=postgres DB_TABLE=bug DB_HOSTNAME=0.0.0.0 DB_PORT=5432 DB_USERNAME=tester DB_PASSWORD=tester go test ./...
  echo "-----"
  DB_DATABASE=postgres DB_TABLE=bug DB_HOSTNAME=0.0.0.0 DB_PORT=5432 DB_USERNAME=tester DB_PASSWORD=tester go test ./... -bench=. -run=$$$
}

if [[ ! -z {1} ]] || [[ "${1}" != "" ]]; then
  ${1}
else
  createDatabase
  sleep 10
  injectStructure
  testCode
fi
