#!/usr/bin/env bash

aws --endpoint-url http://localhost:4572 s3api create-bucket --bucket tester

go build .
zip nest.zip nest
aws --endpoint-url http://localhost:4572 s3 cp ./nest.zip s3://tester/nest.zip

aws --endpoint-url http://localhost:4580 route53 create-hosted-zone --name local.dev --caller-reference devStuff

aws --endpoint-url http://localhost:4581 cloudformation create-stack \
	--template-body file://cf.yaml \
	--stack-name nest \
	--capabilities CAPABILITY_NAMED_IAM \
	--parameters \
		ParameterKey=ServiceName,ParameterValue=nest \
		ParameterKey=BuildKey,ParameterValue=nest.zip \
		ParameterKey=Environment,ParameterValue=dev \
		ParameterKey=BuildBucket,ParameterValue=tester \
		ParameterKey=AuthorizerARN,ParameterValue=tester \
		ParameterKey=CertificateARN,ParameterValue=tester \
		ParameterKey=DNSZoneName,ParameterValue=local.dev \
		ParameterKey=DomainName,ParameterValue=nest.local.dev
