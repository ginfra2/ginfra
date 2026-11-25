#!/usr/bin/env bash

if [ ! -f openapi-generator-cli.jar ]; then
  wget https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/7.2.0/openapi-generator-cli-7.2.0.jar -O openapi-generator-cli.jar
fi

# sed -i -e 's|<doclint>none</doclint>|<doclint>none</doclint><javadocExecutable>${java.home}/bin/javadoc</javadocExecutable>|g'

java -jar openapi-generator-cli.jar generate -g java -i ../../../ginterfaces/service_giconfig.yaml --api-package tech.ginfra.api.service_giconfig.client.api --invoker-package tech.ginfra.api.service_giconfig.client.invoker --model-package tech.ginfra.api.service_giconfig.client.model --package-name tech.ginfra.api.service_giconfig -p "projectName=service_giconfig,artifactId=service_giconfig,artifactVersion=0.0.2,developerEmail=inqueries@ginfra.tech,developerName=ginfra.tech,groupId=tech.ginfra"

grep -q "openapi-generator-cli.jar" .gitignore
if [ $? -gt 0 ] ; then
  echo "openapi-generator-cli.jar" >> .gitignore
fi

rm -rf .github .travis.yml git_push.sh


