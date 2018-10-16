#!/usr/bin/env bash

if [ "$ORG_USER" == "" ]; then
    echo -n "Enter github organization or github account where Kyma is forked:"
    read orgUser
else
    orgUser="$ORG_USER"
fi

if [ -z "$orgUser" ]
then
    echo "ERROR: Github organization or github account is required"
    exit 1
fi

go run ./generator/main.go -template=./../plugins.yaml.tpl -out=./../plugins.yaml -orgUser=${orgUser}

echo "Content of generated file, plugins.yaml:"
cat  ./../plugins.yaml