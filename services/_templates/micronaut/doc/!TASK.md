# Available tasks

### This file will not be updated, so any additions or changes are safe.

### Task command

The [task](https://taskfile.dev/) command that you installed when setting up ginfra automates various build,
test and other processes.  By default, it uses taskfile.yaml find the targets.  You can execute a target simply
by running task in the same directory as the taskfile.yaml giving it the target name.  The following is an example:

```task build```

### Target list

| Task name   | Purpose                                                                                                           |
|-------------|-------------------------------------------------------------------------------------------------------------------|
| build       | Build the service application.  The executable can be found in build/                                             | 
| clean       | Clean the repo and remove any files created from testing.                                                         |
| test        | Run all the tests.                                                                                                |
| docker      | Build the services docker container.  See [Docker container](### Docker container) for more info.                 |
| stubs       | Generate files necessary to generate api stubs.  If you add another api, you should edit taskfile.yaml to add it. | 
| run_local   | Run the service as a local process.                                                                               |

### Docker container

You must have a working docker engine on the same machine you run this.  If you wish to alter the container's name, path
and/or tag, see the file build/dockertask.yaml.

