# userview

This is a simple demonstration app.  It is a Vue/Vite/Typescript application that 
uses a generated openapi client to get user information from the giuser service
.

## Project Setup

```sh
npm install
```

### Compile and Hot-Reload for Development

```sh
npm run dev
```

### Type-Check, Compile and Minify for Production

```sh
npm run build
```

### Run Unit Tests with [Vitest](https://vitest.dev/)

```sh
npm run test:unit+
```

### Run End-to-End Tests with [Nightwatch](https://nightwatchjs.org/)

```sh
# When using CI, the project must be built first.
npm run build

# Runs the end-to-end tests
npm run test:e2e
# Runs the tests only on Chrome
npm run test:e2e -- --env chrome
# Runs the tests of a specific file
npm run test:e2e -- tests/e2e/names.ts
# Runs the tests in debug mode
npm run test:e2e -- --debug
```
    
### Run Headed Component Tests with [Nightwatch Component Testing](https://nightwatchjs.org/guide/component-testing/introduction.html)
  
```sh
npm run test:unit
npm run test:unit -- --headless # for headless testing
```

### Generate stubs

GINFRA_HOME environment variable must be set and the openapi generator is installed.
You can do the latter with the following:
```npm install @openapitools/openapi-generator-cli -g```
WARNING: There are a lot of npm packages called 'openapi' so this can get confused
very easily.  You will usually see an error 'error: unknown option '--input'' if you
have one installed other than the one listed above.  You'll need to clean up your
node packages to get it working.


cd src
cd api
ginfra ginterface merge service_giconfig.yaml
npx openapi --input .\stub.yaml --output giconfig --client axios --name Giconfig
mv stub.yaml giconfig
ginfra ginterface merge service_giuser.yaml
npx openapi --input .\stub.yaml --output giuser --client axios --name Giuser
mv stub.yaml giuser
