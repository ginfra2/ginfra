# Testing

## Unit Testing

These are typical tests that do not require any running application or major component of it.  Unit testing is directly 
supported with go test with the '--tags unit_test' parameter.

## Functional Testing

There are functional tests that do not require running the entire application, but depend on other components within
the application.  They may start a service or run a specific executable if it is directly related to the test area.
Functional testing is directly supported with go test with the '--tags functional_test' parameter.

## Local Testing

Local tests are the same as functional tests, but they will start services, open ports, execute applications (including
Ginfra) and/or other dependencies outside the test itself.  Local testing is directly supported with go test with 
the '--tags local_test' parameter.

Regardless of this of the three types of tests mentioned above, typically they should be implemented in the same
module as they are testing, so that code coverage works properly.

## Higher order testing.

IMPORTANT:  DO NOT put any tests here that participate in 'go test'.  They will be excluded from
coverage if you do.  Tests need to be with the code being tested.  You can put all the tools and 
helpers for those tests here.

Of course, tests that aren't run with 'go test' can go here.

## Next available port

The next available port available for the Minimal Configuration Server is 9502.  If you use that point, please update
this document.

## Ginfra Stub

This is an executable application that provides command line features useful to testing or deployment.


