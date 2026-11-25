$filePath = "openapi-generator-cli.jar"
if (-not (Test-Path -Path $filePath)) {
    Invoke-WebRequest -OutFile openapi-generator-cli.jar https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/7.2.0/openapi-generator-cli-7.2.0.jar
}

java -jar openapi-generator-cli.jar generate -g java -i ..\..\..\ginterfaces\service_giconfig.yaml --api-package tech.ginfra.api.service_giconfig.client.api --invoker-package tech.ginfra.api.service_giconfig.client.invoker --model-package tech.ginfra.api.service_giconfig.client.model --package-name tech.ginfra.api.service_giconfig

$SEL = Select-String -Path .gitignore -Pattern "openapi-generator-cli.jar"
if ($SEL -eq $null)
{
    Add-Content -Path .gitignore -Value "openapi-generator-cli.jar"
}

Remove-Item .travis.yml
Remove-Item git_push.sh
Remove-Item .github -Recurse
