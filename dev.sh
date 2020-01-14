#!/usr/bin/env bash

STACK_NAME=nest
BUILD_BUCKET=builds

export AWS_DEFAULT_REGION=us-east-1

function build()
{
  GOOS=linux GOARCH=amd64 go build .
if [[ ! -f "${STACK_NAME}" ]];then
  echo "build failed"
  exit 1
fi
zip ${STACK_NAME}.zip ${STACK_NAME}
}

function removeFiles()
{
  if [[ -f "${STACK_NAME}.zip" ]]; then
    rm ${STACK_NAME}
    rm ${STACK_NAME}.zip
  fi
}

function moveFiles()
{
  BUCKET_EXISTS=$(awslocal s3api list-buckets | jq '.Buckets[].Name//empty' | grep "${BUILD_BUCKET}")
  if [[ -z "${BUCKET_EXISTS}" ]] || [[ "${BUCKET_EXISTS}" == "" ]]; then
    aws s3api create-bucket --bucket ${BUILD_BUCKET}
  fi
}

removeFiles
build
moveFiles

AUTH_ARN=unknown
AUTH_EXISTS=$(awslocal lambda list-functions | jq '.Functions[].FunctionArn//empty' | grep "authorizer-lambda-dev")
if [[ "${AUTH_EXISTS}" ]] || [[ "${AUTH_EXISTS}" != "" ]]; then
  AUTH_ARN=$(sed -e 's/^"//' -e 's/"$//' <<< ${AUTH_EXISTS})
fi

STACK_EXISTS=$(awslocal cloudformation list-stacks --stack-status-filter ROLLBACK_COMPLETE UPDATE_ROLLBACK_COMPLETE | jq '.StackSummaries[].StackName//empty' | grep "${STACK_NAME}")
if [[ -z "${STACK_EXISTS}" ]] || [[ "${STACK_EXISTS}" == "" ]]; then
  awslocal s3 cp ./${STACK_NAME}.zip s3://${BUILD_BUCKET}/${STACK_NAME}.zip
  awslocal route53 create-hosted-zone --name docker.devel --caller-reference devStuff
  awslocal cloudformation create-stack \
	  --template-body file://cf.yaml \
	  --stack-name ${STACK_NAME} \
	  --capabilities CAPABILITY_NAMED_IAM \
	  --parameters \
		  ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		  ParameterKey=BuildKey,ParameterValue=${STACK_NAME}.zip \
		  ParameterKey=Environment,ParameterValue=dev \
		  ParameterKey=BuildBucket,ParameterValue=${BUILD_BUCKET} \
		  ParameterKey=AuthorizerARN,ParameterValue=${AUTH_ARN} \
		  ParameterKey=CertificateARN,ParameterValue=arn:aws:acm:us-east-1:111122223333:certificate/fb1b9770-a305-495d-aefb-27e5e101ff3 \
		  ParameterKey=DNSZoneName,ParameterValue=docker.devel \
		  ParameterKey=DomainName,ParameterValue=api.docker.devel \
		  ParameterKey=DBEndpoint,ParameterValue=http://10.254.254.254:4569
else
  awslocal s3 cp ./${STACK_NAME}.zip s3://${BUILD_BUCKET}/${STACK_NAME}.zip
  awslocal cloudformation update-stack \
	  --template-body file://cf.yaml \
	  --stack-name ${STACK_NAME} \
	  --capabilities CAPABILITY_NAMED_IAM \
	  --parameters \
		  ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		  ParameterKey=BuildKey,ParameterValue=${STACK_NAME}.zip \
		  ParameterKey=Environment,ParameterValue=dev \
		  ParameterKey=BuildBucket,ParameterValue=${BUILD_BUCKET} \
		  ParameterKey=AuthorizerARN,ParameterValue=${AUTH_ARN} \
		  ParameterKey=CertificateARN,ParameterValue=arn:aws:acm:us-east-1:111122223333:certificate/fb1b9770-a305-495d-aefb-27e5e101ff3 \
		  ParameterKey=DNSZoneName,ParameterValue=docker.devel \
		  ParameterKey=DomainName,ParameterValue=api.docker.devel \
		  ParameterKey=DBEndpoint,ParameterValue=http://10.254.254.254:4569
fi

go test ./...
go test ./... -bench=. -run=$$$

removeFiles
