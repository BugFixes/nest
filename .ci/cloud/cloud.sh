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
  sleep 5
  aws --endpoint-url http://localhost:9324 sqs create-queue --queue-name tester
}

function createStack()
{
  echo "createStack"
  aws cloudformation create-stack \
    --template-body file://.ci/cloud.yaml \
    --stack-name ${STACK_NAME} \
    --capabilities CAPABILITY_NAMED_IAM \
    --parameters \
      ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
      ParameterKey=Environment,ParameterValue=live \
      ParameterKey=BuildBucket,ParameterValue=${BUILD_BUCKET} \
      ParameterKey=BuildKey,ParameterValue=${BUILD_KEY} \
      ParameterKey=DBHostname,ParameterValue=${DB_HOSTNAME} \
      ParameterKey=DBPort,ParameterValue=${DB_PORT} \
      ParameterKey=DBUsername,ParameterValue=${DB_USERNAME} \
      ParameterKey=DBPassword,ParameterValue=${DB_PASSWORD} \
      ParameterKey=DBDatabase,ParameterValue=${DB_DATABASE} \
      ParameterKey=DBTable,ParameterValue=${DB_TABLE} \
      ParameterKey=SQSHostname,ParameterValue=${SQS_HOSTNAME} \
      ParameterKey=SQSQueueName,ParameterValue=${SQS_QUEUENAME}
}

function deleteStack()
{
  echo "deleteStack"
  aws cloudformation delete-stack --stack-name ${STACK_NAME}
}

function updateStack()
{
  echo "updateStack"
  aws cloudformation update-stack \
    --template-body file://.ci/cloud.yaml \
    --stack-name ${STACK_NAME} \
    --capabilities CAPABILITY_NAMED_IAM \
    --parameters \
      ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
      ParameterKey=Environment,ParameterValue=live \
      ParameterKey=BuildBucket,ParameterValue=${BUILD_BUCKET} \
      ParameterKey=BuildKey,ParameterValue=${BUILD_KEY} \
      ParameterKey=DBHostname,ParameterValue=${DB_HOSTNAME} \
      ParameterKey=DBPort,ParameterValue=${DB_PORT} \
      ParameterKey=DBUsername,ParameterValue=${DB_USERNAME} \
      ParameterKey=DBPassword,ParameterValue=${DB_PASSWORD} \
      ParameterKey=DBDatabase,ParameterValue=${DB_DATABASE} \
      ParameterKey=DBTable,ParameterValue=${DB_TABLE} \
      ParameterKey=SQSHostname,ParameterValue=${SQS_HOSTNAME} \
      ParameterKey=SQSQueueName,ParameterValue=${SQS_QUEUENAME}
}

function cloudFormation()
{
  echo "cloudFormation"
}

function testIt()
{
  echo "testIt"
  DB_DATABASE=postgres \
  DB_TABLE=bug \
  DB_HOSTNAME=0.0.0.0 \
  DB_PORT=5432 \
  DB_USERNAME=postgres \
  DB_PASSWORD=tester \
  SQS_REGION=eu-west-2 \
  SQS_QUEUE=tester \
  SQS_ENDPOINT=http://localhost:9324 \
  go test ./...
}

if [[ ! -z ${1} ]] || [[ "${1}" != "" ]]; then
  ${1}
else
  build
  moveFiles
  cloudFormation
fi

