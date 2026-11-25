# Service_gipet

This file will not be updated, so it is safe to make any changes.

### Service information

The gipet service is a very simple demonstration of Micronaut.  Through openapi you can ask it for a pet type
and it will tell you what breed you get.

## Specifics

[Specifics](doc/Specifics.md) is the mechanism where the service can tell GInfra information about how it should be run.
See the [additional documentation](doc/Specifics) for more information.

## File layout
| Directory/File      | Purpose                                                                              |
|---------------------|--------------------------------------------------------------------------------------|
| api/                | Generated service stubs and code.                                                    | 
| assets/             | Target build directory and container build files.                                    |
| build.gradle        | The gradle build configuration                                                       |
| doc/                | Additional documentation.                                                            |                                                                           |
| doc/                | Additional documentation.                                                            |                                                                           |
| gradle/             | The gradle wrapper.  LEt it alone.                                                   |
| gradlew             | Sh script to run gradle (in the wrapper).                                            |
| gradle.bat          | Windows BAT script to run gradle (in the wrapper).                                   |
| log/                | Log files from runtime.   They can be removed whenever (but lead the README.md file) |
| micronaut-cli.yml   | Configuration for the micronaut cli.  There are know specific uses for it as is.     | 
| settings.gradle     | Settings file for gradle                                                             |
| src/                | Source code for the service and tests                                                | 
| test/               | Data and configuration for local testing.                                            |
| service_config.json | Dynamic configuration file.  This one is provided purely for testing purposes.       |
| start_local.cmd     | Start the service as a local executable (windows)                                    |
| start_local.sh      | Start the service as a local executable (linux/osx/etc)                              |
| taskfile.yaml       | Default task target file.  See below.  DO NOT make changes to this one.              |
| taskfilecustom.yaml | Custom task target file.  You may make changes to this one.                          |


### Note about start_local.*

While you can run these yourself, they are meant to be started with the 'ginfra manage local
start' command.  It will handle any dependencies and configuration.

### Developer information

For automated tasks that build and do other actions, see the [task](doc/TASK.md) document.

