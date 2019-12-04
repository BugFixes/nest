#!/usr/bin/env bash

STACK_NAME=nest

rm ${STACK_NAME}
rm ${STACK_NAME}.zip

GOOS=linux GOARCH=amd64 go build .
zip ${STACK_NAME}.zip ${STACK_NAME}

STACK_EXISTS=$(aws --region us-east-1 --endpoint-url http://localhost:4581 cloudformation list-stacks --stack-status-filter ROLLBACK_COMPLETE UPDATE_ROLLBACK_COMPLETE | jq '.StackSummaries[].StackName//empty' | grep "${STACK_NAME}")
if [[ -z "$STACK_EXISTS" ]] || [[ "$STACK_EXISTS" == "" ]]; then
  aws --region us-east-1 --endpoint-url http://localhost:4572 s3api create-bucket --bucket tester
  aws --region us-east-1 --endpoint-url http://localhost:4572 s3 cp ./${STACK_NAME}.zip s3://tester/${STACK_NAME}.zip
  aws --region us-east-1 --endpoint-url http://localhost:4580 route53 create-hosted-zone --name docker.devel --caller-reference devStuff
  aws --region us-east-1 --endpoint-url http://localhost:4581 cloudformation create-stack \
	  --template-body file://cf.yaml \
	  --stack-name ${STACK_NAME} \
	  --capabilities CAPABILITY_NAMED_IAM \
	  --parameters \
		  ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		  ParameterKey=BuildKey,ParameterValue=${STACK_NAME}.zip \
		  ParameterKey=Environment,ParameterValue=dev \
		  ParameterKey=BuildBucket,ParameterValue=tester \
		  ParameterKey=AuthorizerARN,ParameterValue=tester \
		  ParameterKey=CertificateARN,ParameterValue=tester \
		  ParameterKey=DNSZoneName,ParameterValue=docker.devel \
		  ParameterKey=DomainName,ParameterValue=api.docker.devel
else
  aws --region us-east-1 --endpoint-url http://localhost:4572 s3 cp ./${STACK_NAME}.zip s3://tester/${STACK_NAME}.zip
  aws --region us-east-1 --endpoint-url http://localhost:4581 cloudformation update-stack \
	  --template-body file://cf.yaml \
	  --stack-name ${STACK_NAME} \
	  --capabilities CAPABILITY_NAMED_IAM \
	  --parameters \
		  ParameterKey=ServiceName,ParameterValue=${STACK_NAME} \
		  ParameterKey=BuildKey,ParameterValue=${STACK_NAME}.zip \
		  ParameterKey=Environment,ParameterValue=dev \
		  ParameterKey=BuildBucket,ParameterValue=tester \
		  ParameterKey=AuthorizerARN,ParameterValue=tester \
		  ParameterKey=CertificateARN,ParameterValue=tester \
		  ParameterKey=DNSZoneName,ParameterValue=docker.devel \
		  ParameterKey=DomainName,ParameterValue=api.docker.devel
fi
