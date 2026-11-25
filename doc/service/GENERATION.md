## Generation feature of Ginfra App

### Generate a service.

It is easy to generate a service framework where all you need to do is add functionality.
1- Create the directory where you want it to go.
2- Change directory into there.
3- Run the generate command.

The command looks like the following:

```ginfra generate service [Service name] [Service type] [Module Name] ```

Service name is the name for your service.
Service type is a supported class of service.  Currently, there are four provided by ginfra:

|    Type    | What it does                                                                      |
|:----------:|-----------------------------------------------------------------------------------|
|  apiecho   | An [Echo](https://echo.labstack.com/) based API Go service.                       |
|   server   | A generic server layout, where you provide the server (i.e. postgres, mongo, etc) |
| viewvitets | A [Vue.js](https://vuejs.org/) web application service.                           | 
| micronaut  | A [Micronaut](https://micronaut.io/) based java service                           |

The module name for your service.  This isn't always used, but would be important for Go (and other similar languages), 
if you were to be using that.  'apiecho' requires this is a value Go module name.

The following is an example of a generate command:

```ginfra generate service myservice apiecho myrepo.github.com/me/myservice```

### Custom templates

You can use your own service templates, just specify the template directory in the command line with the '--template'
parameter.  The generator will look for a subdirectory in the template directory that is the same as the 
[Service type] and treat its contents as the service template.



