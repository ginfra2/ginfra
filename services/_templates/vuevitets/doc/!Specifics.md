# Specifics

This file will not be updated, so any additions or changes are safe.

This is the additional information available to Ginfra on how to run the service.

## Configuration 

Specifics are contained in the file [specifics.json][../specifics.json].  The file must
be present at runtime but may be an empty json (simple a "{}".)

## Sections

### Exposed

Exposed tells Ginfra what ports it wished to be exposed.  Each entry represents a port.
The key will be the port value as a string.  Within the entry, all that is supported at
this time is a "name" key with an arbitrary value.


