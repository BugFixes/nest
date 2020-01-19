#!/usr/bin/env bash

BUILD_BUCKET=bugfixes-builds-eu
STACK_NAME=nest

function build()
{
  echo "build"
  GOOS=linux GOARCH=amd64 go build .
  zip ${STACK_NAME}-${GITHUB_SHA}.zip ${STACK_NAME}
}

function moveFiles()
{
  echo "moveFiles"
  aws s3 cp ./${STACK_NAME}-${GITHUB_SHA}.zip s3://${BUILD_BUCKET}/${STACK_NAME}-${GITHUB_SHA}.zip
}

function setupDatabase()
{
  echo "setupDatabase"

  docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=tester -e POSTGRES_USERNAME=postgres -e POSTGRES_DB=postgres --name tester_postgres postgres:11.5
  sleep 5
  docker exec -e PGPASSWORD=tester tester_postgres psql -U postgres -c "CREATE TABLE "public"."agent" ("id" uuid, "name" varchar(200), "key" uuid, "secret" uuid, "company_id" uuid, PRIMARY KEY("id"));"
  docker exec -e PGPASSWORD=tester tester_postgres psql -U postgres -c "CREATE TABLE "public"."bug" ("id" uuid, "hash" text, "message" text, "agent_id" uuid, "level" int4, "time_posted" timestamp, PRIMARY KEY("id"));"
}

function setupQueue()
{
  echo "setupQueue"
  docker run -d -p 9324:9324 -p 9325:9325 --name tester_queue roribio16/alpine-sqs:latest
  aws --endpoint-url http://localhost:9324 sqs create-queue --queue-name bugfixes-hive
}

function createStack()
{
  echo "createStack"
}

function deleteStack()
{
  echo "deleteStack"
  aws cloudformation delete-stack --stack-name ${STACK_NAME}
}

function updateStack()
{
  echo "updateStack"
}

function cloudFormation()
{
  echo "cloudFormation"
}

function testIt()
{
  echo "testIt"
}

if [[ ! -z ${1} ]] || [[ "${1}" != "" ]]; then
  ${1}
else
  build
  moveFiles
  cloudFormation
fi

