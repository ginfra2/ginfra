# Service implementation

### This file will not be updated, so any additions or changes are safe.

## Service information

... custom to your service...

## Specifics

[Specifics](doc/Specifics.md) is the mechanism where the service can tell GInfra information about how it should be run.
See the [additional documentation](doc/Specifics) for more information.

## File layout
| Directory/File      | Purpose                                                                                                                           |
|---------------------|-----------------------------------------------------------------------------------------------------------------------------------|
| api/                | Generated service stubs and code.                                                                                                 | 
| build/              | Target build directory and container build files.                                                                                 |
| doc/                | Additional documentation.                                                                                                         |
| local/              | Data, values and tools specific to your service.                                                                                  |
| providers/          | Service providers that are unique to your service.                                                                                |
| service/            | Your service implementation.  This is where you will implement the interfaces defined in your API specification and tests for it. |
| service/private.go  | Server overhead.  Generally you can leave it alone.                                                                               |
| test/               | Data and configuration for local testing.                                                                                         |
| service_config.json | Dynamic configuration file.  This one is provided purely for testing purposes.                                                    |
| start_local.cmd     | Start the service as a local executable (windows)                                                                                 |
| start_local.sh      | Start the service as a local executable (linux/osx/etc)                                                                           |
| taskfile.yaml       | Default task target file.  See below.                                                                                             |
| taskfilecustom.yaml | Custom task target file.  You may make changes to this one.                                                                       |

### Note about start_local.*

While you can run these yourself, they are meant to be started with the 'ginfra manage local
start' command.  It will handle any dependencies and configuration.

## Developer information

For automated tasks that build and do other actions, see the [task](doc/TASK.md) document.



