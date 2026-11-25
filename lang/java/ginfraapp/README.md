# GINFRA java application

Exposes from of the ginfra java library on the command line.

## Setup

The following are prerequisites:
- You have installed the GINFRA application.
- You have installed maven.  In debian/ubuntu systems, you can do this with: sudo apt install maven
- You have installed JDK version 1.17 and it is in the PATH.  Later versions may work, but they could conflict with gradle/maven in other GINFRA projects.

## Building

You can use the typical gradle tasks, but when you are ready to push to your local maven repository, run the 
following command:

```task release```

## Using

If you are using gradle, add it to your dependencies.

dependencies {
    implementation("tech.ginfra:ginfra:1.0.0")
}