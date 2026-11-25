# Service implementation

This is a very minimal implementation to support server_control.

### Service information

The user service is a very simple user information provider.  It is NOT an authentication or 
authorization service.  

### File layout

| Directory/File      | Purpose                                                                                                                           |
|---------------------|-----------------------------------------------------------------------------------------------------------------------------------|
| api/                | Generated service stubs and code.                                                                                                 | 
| local/              | Data, values and tools specific to your service.                                                                                  |
| providers/          | Service providers that are unique to your service.                                                                                |
| service/            | Your service implementation.  This is where you will implement the interfaces defined in your API specification and tests for it. |
| service/private.go  | Server overhead.  Generally you can leave it alone.                                                                               |

### Developer information

For automated tasks that build and do other actions, see the [task](doc/TASK.md) document.
