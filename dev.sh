#!/usr/bin/env bash

STACK_NAME=nest
BUILD_BUCKET=builds

removeFiles()
{
  if [[ -f "${STACK_NAME}.zip" ]]; then
    rm ${STACK_NAME}
    rm ${STACK_NAME}.zip
  fi
}

removeFiles

GOOS=linux GOARCH=amd64 go build .
if [[ ! -f "${STACK_NAME}" ]];then
  exit "build failed"
fi
zip ${STACK_NAME}.zip ${STACK_NAME}

BUCKET_EXISTS=$(aws --region us-east-1 --endpoint-url http://localhost:4572 s3api list-buckets | jq '.Buckets[].Name//empty' | grep "${BUILD_BUCKET}")
if [[ -z "${BUCKET_EXSISTS}" ]] || [[ "${BUCKET_EXISTS}" == "" ]]; then
  aws --region us-east-1 --endpoint-url http://localhost:4572 s3api create-bucket --bucket ${BUILD_BUCKET}
fi

AUTH_ARN=unknown
AUTH_EXISTS=$(aws --region us-east-1 --endpoint-url http://localhost:4574 lambda list-functions | jq '.Functions[].FunctionArn//empty' | grep "authorizer-lambda-dev")
if [[ "${AUTH_EXISTS}" ]] || [[ "${AUTH_EXISTS}" != "" ]]; then	
  AUTH_ARN=$(sed -e 's/^"//' -e 's/"$//' <<< ${AUTH_EXISTS})
fi

STACK_EXISTS=$(aws --region us-east-1 --endpoint-url http://localhost:4581 cloudformation list-stacks --stack-status-filter ROLLBACK_COMPLETE UPDATE_ROLLBACK_COMPLETE | jq '.StackSummaries[].StackName//empty' | grep "${STACK_NAME}")
if [[ -z "${STACK_EXISTS}" ]] || [[ "${STACK_EXISTS}" == "" ]]; then
  aws --region us-east-1 --endpoint-url http://localhost:4572 s3 cp ./${STACK_NAME}.zip s3://${BUILD_BUCKET}/${STACK_NAME}.zip
  aws --region us-east-1 --endpoint-url http://localhost:4580 route53 create-hosted-zone --name docker.devel --caller-reference devStuff
  aws --region us-east-1 --endpoint-url http://localhost:4581 cloudformation create-stack \
	  --template-body file://cf.yaml \
	  --stack-name ${STACK_NAME} \
	  --capabilities CAPABILITY_NAMED_IAM \
	  --parameters \
		  ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		  ParameterKey=BuildKey,ParameterValue=${STACK_NAME}.zip \
		  ParameterKey=Environment,ParameterValue=dev \
		  ParameterKey=BuildBucket,ParameterValue=${BUILD_BUCKET} \
		  ParameterKey=AuthorizerARN,ParameterValue=${AUTH_ARN} \
		  ParameterKey=CertificateARN,ParameterValue=tester \
		  ParameterKey=DNSZoneName,ParameterValue=docker.devel \
		  ParameterKey=DomainName,ParameterValue=api.docker.devel
else
  aws --region us-east-1 --endpoint-url http://localhost:4572 s3 cp ./${STACK_NAME}.zip s3://${BUILD_BUCKET}/${STACK_NAME}.zip
  aws --region us-east-1 --endpoint-url http://localhost:4581 cloudformation update-stack \
	  --template-body file://cf.yaml \
	  --stack-name ${STACK_NAME} \
	  --capabilities CAPABILITY_NAMED_IAM \
	  --parameters \
		  ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		  ParameterKey=BuildKey,ParameterValue=${STACK_NAME}.zip \
		  ParameterKey=Environment,ParameterValue=dev \
		  ParameterKey=BuildBucket,ParameterValue=${BUILD_BUCKET} \
		  ParameterKey=AuthorizerARN,ParameterValue=${AUTH_ARN} \
		  ParameterKey=CertificateARN,ParameterValue=tester \
		  ParameterKey=DNSZoneName,ParameterValue=docker.devel \
		  ParameterKey=DomainName,ParameterValue=api.docker.devel
fi

go test ./...
go test ./... -bench=.

removeFiles
