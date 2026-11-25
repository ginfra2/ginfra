# JAVA

## Prerequisites

To build and use the java modules, you need to install the following:

- JAVA SDK : Install [java](sudo apt install openjdk-17) and make sure it is in your PATH.  It works best with version 17.  
  If you go later, you may end up with gradle incompatibility issues.   On ubuntu, I successfully installed the jdk with
  the following command.   ``` apt install -y openjdk-17-jdk ```
- Maven : Install [maven](https://maven.apache.org/).  It is not important how you install it, but I found it very easy
  to install on Ubuntu and Debian with the following command: ```apt install -y maven```

## Build, package and push 

To build, package and (in some cases) push the modules into your local maven repo, run the following from the
repo root.

```
task java_push
```

This will be necessary if you plan on building certain gservices things such as the newdemo or grand demo.

## Packages/Modules

- ginfra : Common ginfra functionality in a library like found in the ginfra go common/ module.
- giconfig : A configuration provider client. 
- ginfrapp : Exposes ginfra java library functions on the command line.

## Usage

For now, the best thing to do is look at the example java service [gipet](https://gitlab.com/ginfra/gservices/-/tree/main/service_gipet?ref_type=heads).
