#!/usr/bin/env bash
aws dynamodb delete-table \
	--table-name bugs \
	--endpoint-url http://docker.devel:4569

aws dynamodb create-table \
	--table-name bugs \
	--endpoint-url http://docker.devel:4569 \
	--attribute-definitions AttributeName=identifier,AttributeType=S \
	--provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5 \
	--key-schema AttributeName=identifier,KeyType=HASH
