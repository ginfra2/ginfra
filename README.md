# ginfra

## TL;DR

This is an application, libraries, services and tools for cloud based development, testing and hosting with an emphasis 
on easy and fast use.

_**Now with AI driven testing!**_

## What is this?

As a developer, did you ever wish you could develop and run your stuff on your laptop and know that without 
any additional work, it can run in production?  You don't even want to know what happens in production.
As a production operator, have you ever wished you could use a simple tool with straight forward configuration 
to run them?   Yes?  Then you have come to the right place.  That is the goal of the Ginfra project.  

As of now, I have provided an example of how a developer can run a complete end to end on his laptop
while knowing very little about what happens behind the scenes?  And hey, if you do want to know, it is all
in the source code here.

So far, it will do this:
- Manage OpenApi specifications and make golang stubs for you, with support for other languages.
- Create an entire service framework that will use those Apis leaving you to only implement that entrypoints (but you can 
  do more too).
- Manage system/cluster/cloud wide configuration.
- Enable testing of the service locally.
- Run and test a service as process on your laptop/desktop.
- Run and test the service in docker containers or kubernates.
- Provide a configuration service, as well a demonstration user information service and a demonstration website for a 
  complete end-to-end.
- Provide orchestration scripting for preproduction operation (development, test, etc).
- Support for OpenTelemetry, Selenium, Ollama and other technologies.
- Support for AI driven testing through test service framework.
- Service generation from templates supporting EchoApi (REST), VueVite application, general server host and NodeJs, or 
  any templates you create.

While ginfra is implemented in Go, your project doesn't need to be in Go to take advantage of it.  One of the demo
services is written in java and the demo web apps are typescript.

If you are looking for best-practices, the project isn't mature enough for that yet.  For instance, only http is used 
and the configuration service uses a file storage.  I'll get there, but the goal right now is to show off the guiding
principle--developers shouldn't need to know what operations does and operations shouldn't need to know what developers 
do.  Make it easy to do both...

... and then we will work to keep the hackers out.

## What are the moving parts?

### Ginfra application

There are only one thing you have to care about.  The ginfra app.  Getting the ginfra app is super easy; just
do this:

- Have a working [go](https://go.dev/) developer environment.  Note that your go bin directory should be in your PATH.  It 
  is usually go/bin/ in your home directory.  If the directory does not exist yet, run the following
  command to install task and then add the directory to your PATH.
- Install [task](https://taskfile.dev/) (if you don't have it already).  One way to do that is just to run this 
  command: go install github.com/go-task/task/v3/cmd/task@latest.  If task does not run, it is either because
  GOBIN is not set or the go build binaries are not in the path.  On a typical linux installation, the binaries
  will be put in ~/go/bin, so you could just add this to the path.  
- Clone this repo:  [ginfra/ginfra](https://gitlab.com/ginfra/ginfra)  To get the latest tested version:  
```git clone https://gitlab.com/ginfra/ginfra.git; cd ginfra```
- cd into ginfra/.
- Add an environment variable called GINFRA_HOME and point it at the repo.  Remember to close/restart your terminals 
  after doing this!  (In some cases, such as certain desktop linux, you may have to reboot.)
- Run this command: task install_notest

But shucks, you don't have linux?  Just Windows?  Good news, [it works on windows too](https://gitlab.com/ginfra/ginfra/-/blob/main/doc/WHYWINDOWS.md).

After this, you should be able to run 'ginfra' on the command line and get some help.  There is some
[additional documentation](doc/cmd/README.md) available too.

### Ginfra libraries

For the log language, common/ module provides a lot of useful functionality for apps and services.  
In fact, the ginfra  app uses it.  But good news: you don't ever actually have to look at it if you don't want to dd so.
The ginfra app is able to create and maintain a service framework for you and hide tons of the complexity for 
running it.  It uses the common/modules, so you don't have to.

Some of this functionality (as needed to support a service) is [also available in java](lang/java/README.md).

### Gservices

The [ginfra/gservices](https://gitlab.com/ginfra/gservices) is a model for how to use ginfra.
Typically, you would clone that repo next to the ginfra, and have the same version checked out as Ginfra.
There are release versions, but main branch is the latest tested version.  It is up to how you plan
on developing your services whether you should stay with main or freeze to a branch.  So far, no updates have broken 
anything and we plan to keep it that way.

### End to end demonstration

[Ginfra/gservices](https://gitlab.com/ginfra/gservices) contains and end to end demonstration of how all this works.  Check out [the documentation](https://gitlab.com/ginfra/gservices/-/blob/main/integration/demo/README.md) 
on how to run it.

There is also a more complicated end to end demonstration.  Check out [the documentation](https://gitlab.com/ginfra/gservices/-/blob/main/integration/scenerios/newdemo/README.md).

There is a grand end to end demonstration that includes a more complete environment, including an observability solution.  
However, it may be a little much for some laptops.   This demo is for host mode only, which requires a docker host to run.   
Check out [the documentation](https://gitlab.com/ginfra/gservices/-/blob/main/integration/scenerios/grand/README.md).

There is a simple [end to end demo](https://gitlab.com/ginfra/gservices/-/blob/main/integration/demo/kube/README.md) that 
is hosted on kubernetes.  It does require a working kubernetes cluster, but I do [provide a way to build one](https://gitlab.com/ginfra/gplatform/-/blob/main/platform/nix_root/kube/README.md?ref_type=heads). 

There is smaller demo involving observability. This demo is for host mode only, which requires a docker host to run.  
See the [otel demo](https://gitlab.com/ginfra/gservices/-/tree/main/integration/workbench/otel/host) for more information.

(NEW!)  Finally, there is a [QA/AI demo](https://gitlab.com/ginfra/gservices/-/blob/main/integration/scenerios/demoqa/README.md) will show an application end to end and the ability 
to test it using AI-driven tools.  The support is PRELIMINARY (ok, maybe closer to BETA) but will see active development.

## Roadmap

Most of my original vision has been completed.  What you don't see is the forked repo with the EKS support, but that belongs
to a company I've been working with, so I can't show it here.  However, we do now have a complete Kubernetes end to end.

Here is a loose roadmap of upcoming work...
- UNDERWAY test service with controller UI using magnitude and playwright.
- TODO register external services for management and participation.  Include register local service into host cloud for testing.
- TODO Edge specific service or service configuration.  1) nginx--api and app proxy.  2) edge collector for telemetry.
- ~~PHASE 2 DONE - AI based testing framework.  Magnitude test driver running in test service.~~  ~~Ollama model host complete with demonstration application.~~ 
- CANCELED FOR NOW ~~[OAUTH](https://oauth.net/2/) for services and web.~~
- DONE ~~Additional service generation templates, including [Vue.js](https://oauth.net/2/) app, and generic Node.js~~.
- PENDING (waiting feedback) - Python service framework and demo application.
- DONE ~~First managed service using kubernetes.~~
- DONE Part 1) ~~Portal/Ingress service Part 2) with TLS.  Went with bundle" concept instead by request.~~
- DONE (though docs can always improve) ~~Improve documentation.~~
- DONE ~~[OpenTelemtry](https://opentelemetry.io/) support.~~
- DONE ~~Include Java/Micronaut as a template service framework.~~
- DONE ~~Greatly expand testing and get CI pipeline running in gitlab.~~
- DONE ~~Simple orchestration scripting for non-production uses.~~
- DONE ~~Complete end-to-end with Go service, Java service, web frontend and observability platform with UI.~~


## Additional install information.

If you plan on running code coverage, which is part of the 'ci' task, you need to install gotestsum.  You can do so with the 
following command:

```go install gotest.tools/gotestsum@latest```

## Additional information

- [Ginfra application documentation](doc/cmd/README.md)
- [Developing for Ginfra](doc/DEVELOPMENT.md)
- [About Services and developing them](doc/service/README.md)
- [Service generation using Ginfra](doc/service/GENERATION.md)
- [Document Index](doc/README.md)
- [LICENSE](LICENSE)

... but what is TillieAI?  It is a working AI chat application, but it was built to prank someone.  See if you can find 
how the prank worked....
