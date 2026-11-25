# Notes on developing Ginfra

## Repo map

What you find in this repository.

|                            Dir/File                            | What it does                                          |
|:--------------------------------------------------------------:|-------------------------------------------------------|
|                    [api/](../api/README.md)                    | Generated API stubs used by the ginfra app.           |
|                    [app/](../app/README.md)                    | The Ginfra application code.                          |
|                  [build/](../build/README.md)                  | Build target, assets and tools                        |
|                 [common/](../common/README.md)                 | Common code used by the app and services (if desired) |
|                    [doc/](../doc/README.md)                    | Documentation.                                        |
|            [ginterfaces/](../ginterfaces/README.md)            | APIs defined by Ginfra.                               |
|                   [lang/](../lang/README.md)                   | Support for languages over than Go                    |
|             [services/](../ginterfaces/README.md)              | Services and service templates provided by Ginfra.    |
|                   [test/](../test/README.md)                   | Test information, assets and implementations.         |
|                              tmp/                              | Directory for temporary files mostly used in testing. |
| [.gitlab-ci.yml](https://gitlab.com/ginfra/ginfra/-/pipelines) | Gitlab CI pipeline definitions file.                  |
|                         taskfile.yaml                          | Build task file.  See [Builds](##Builds) below        |
|                       taskfile-lang.yaml                       | Support for building other languages                  |

## Builds

Ginfra uses the [Task](https://taskfile.dev/) tool to do builds.  All tasks are run by a
simple 'task [target]' where the [target] is one of below.

|     Target     | What it does                                                                       |
|:--------------:|------------------------------------------------------------------------------------|
|    install     | Build, test and install the ginfra application.                                    |
| install_notest | Build and install without testing.                                                 |
|     build      | Build only.                                                                        |
|      test      | Build and test only.  (Unit and Functional tests)                                  |
|     clean      | Clean the environment and go modules.  Do this is builds are inexplicably failing. |
|    testall     | Run all available tests.                                                           |
|       ci       | Run the CI processing which includes build, all tests and code coverage report.    |







