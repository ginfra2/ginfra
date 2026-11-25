# GINFA COMMAND : RUN

## RUN

ginfra run [run script]

[run script] is script to run.

## RUN SCRIPTS

The scripts are yaml meant to help with simple service orchestration.  It is not meant to replace tools used in 
production, CI environments or whatever.  

The following is an example of a script:

```
---
Comment:
  - Info: General stuff.
Meta:
  Comment: Stuff
  Deployment: thisdeployment
Workflow:
  - Ginfra:
      - WorkingDir: ../../../service_giconfig
      - Remove: log/*.log
      - Detach: true
      - Cmd: manage local start config:// --clear
      - Output:
        - Delay: 300
        - Until: 4000
        - Source: log/ginfra.log
        - Scan: Service starting
      - Ping:
          Delay: 50
          Until: 6000
          Url: http://localhost:5111/specifics
          CodeExpect: 200
      - Success: Configuration provider starting.
  - Log: Done
```

There are three top level directives: 'Comment', 'Meta' and 'Workflow'

Note that most directives are optional, but if no actual actions or steps are specified, nothing may happen.

### Directive: Comment

Comments are unstructured yaml objects that can contain anything and while be completely ignored when processed.

### Directive: Log

The given string will be logged after subjected to string replacement.

### Directive: Meta

Meta contains configurations that apply to all actions in the script.  It can have the following directives: 'Comment',
'WorkingDir', 'Deployment', 'Terminate', and 'Config'.  WorkingDir can be overridden by a similar directive in a workflow 
object.

### Directive: Workflow

These are the sequential workflow steps in a yaml list (array).  There are four types of action objects that can be in this 
list: 'Cmd', 'Ginfra', 'Comment' or 'Log'.  The Comment will be ignored, like any Comment directive.  Cmd specifies a command
line command to execute.  Ginfra will run the ginfra application.  Both these objects will have a 'Cmd' entry which 
will specify the command and parameters (you will omit 'ginfra' from this entry when used in Ginfra objects.)

Cmd objects can contain the following directives: 'Platform', 'WorkingDir', 'Remove', 'FailOk', 'Cmd' (as described above), 
'Post', 'Output', 'Ping', 'Success' and 'Comment'.

Ginfra objects can contain the following directives: 'WorkingDir', 'Remove', 'Cmd' (as described above), 'Post', 'Detach',
'Output', 'Ping', 'Success' and 'Comment'.

If a Ginfra results in a Hosting.json file appearing (such as 'ginfra host init' or 'ginfra kube init'), the file will
be loaded along with the specified configuration provider client.  See the Variable replacement section below for more
information on what this provides.

Workflow action objects are considered successful if the command execute is successful (exit 0) and any Output and Ping
directives are also successful.  If any action object is not successful, the script is halted, unless the FailOk 
directive is used.

### Directive: Deployment

This directive specifies the deployment for any Ginfra workflow action objects.

### Directive: Terminate

If set to true, it will terminate all detached workflow actions when the end of the script is reached.  If not set to true,
the script will pause and wait for either the detached actions to finish or the user interrupts the script (such as with
a kill or Ctrl-C).

### Directive: Config

This directive specifies a file containing name value pairs that are separated by '='.  All other directives in the script
will be subject to variable replacement, where the names are contained in  <<< >>>.  See [this example of 
the configuration file](../../app/run/test/3.config) and how it is used in [a script](../../app/run/test/3.yaml).  If multiple 
Config directives are given, only the last one will be used.

### Directive: Platform

This specifies the platform on which to run the workflow action objects.  The following is an example:

```
---
Workflow:
  - Cmd:
      Platform: windows
      Cmd: ginfra -v
      Post: t
      Output:
        Scan: version
        NoScan: management
      Success: Windows command is present.
  - Cmd:
      Platform: nix
      Cmd: ginfra -v
      Post: t
      Success: Unix command is present.
```

The value for Platform may be 'windows' or 'nix'.  In the above example, only the first workflow action will be run if
it is executed in a windows environment.  Only the second workflow action will be run if it is executed in a unix environment.
A unix environment includes Linux, OSX and any other comment unix flavor (though it has only been tested on the first two).

### Directive: WorkingDir

Sets the working directory for the workflow actions. If in the Meta directive, it is the default for all the actions.
Each action can then override it.

### Directive: 'Remove'

It will remove the files at the path given.  Wildcards can be used.  It will silently ignore any errors, so be mindful
of file permissions.

### Directive: 'FailOk'

Normally, any failed action object (defined as exiting with other than 0) will cause the script to halt.  IF this directive
is set as true, the failure will be ignored and the script will continue.

### Directive: 'Post'

If this directive is set to true, any output from a workflow action object will be printed.  Otherwise, the output 
is quietly ignored.

### Directive: 'Detach'

If this directive is set to true, the workflow action will be detached to run in the background and the script will 
continue to execute.  Normally, the script will pause until the action is complete.

### Directive: 'Output'

```
      - Output:
        - Delay: 100
        - Until: 4000
        - Scan: DeploymentName
        - Source: Hosting.json
      - Success: Configuration provider started.
```

This directive is for verifying output from the output from a workflow action.  The sub-directives can be given 
as a list (yaml array) and will be executed in that order.  It is possible to give it as an object, but the execution
would be random and thus should generally be avoided.  For instance, if you put a Scan before a Source, the scan will
be done on the command output rather than the source file.  Sub-directives can be repeated as much as desired. 

It can have the following sub-directives:

- Source : This is a file to Scan, NoScan, Extract and Capture.  If Source is not given, the command output will be used.
- Scan : The source will be scanned for the text given.  If the text is found, this is considered successful.
- NoScan : The source will be scanned for the text given.  If the text is NOT found, this is considered successful.
- Delay : The time in milliseconds to wait until starting the Scan and/or NoScan.
- Until : The time in milliseconds in which any Source, Scan and NoScan should be repeated until successful.  After the time given, these operations will be left to fail.  It can be renewed with additional Until directives.
- Terminate : If the workflow action is detached, at the completion of any Scan and/or NoScan, the action will be terminated.
- Comment : This is ignored like any Comment directive.
- Extract : It will extract text from Source or output using the given RegEx.  The result will be put in the variable specified in Target.
- Capture : It will capture all the text from Source or output.  The result will be put in the variable specified in Target.

Note that whitespace is significant in yaml.  In the examples below you will see the most common mistake
with spacing.  The object items under Extract and Capture need to be spaced beyond the list entry or they will be 
considered new list entries.

Bad
```
      Output:
        - Delay: 1000
        - Extract:
          RegEx: H[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]
          Target: SnippetId
```

Good
```
      Output:
        - Delay: 1000
        - Extract:
            RegEx: H[0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9][0-9]
            Target: SnippetId
```


### Directive: 'Ping'

```
      Ping:
        Delay: 50
        Until: 6000
        Url: http://localhost:8900/specifics
        CodeExpect: 200
```

This directive is for verifying connectivity to any process started by a workflow action.  It will attempt an http
get from the URL given.  It can have the following sub-directives:

- URL : The URL to ping.
- Delay : The time in milliseconds to wait until starting the ping.
- Until : The maximum time to wait for it to connect.  The ping will be periodically attempted until then.
- CodeExpect : The HTTP code expected from the server.  If other codes are received, it will retry until it gets the code or the Until time is reached.

This directive is considered successful if the CodeExpect is received.  If the CodeExpect is not specified,
any connectivity is considered successful.

### Directive: 'Repeat'

Detached commands cannot be repeated.

### Directive: 'Success'

If the workflow action is successful, the text given will be printed. 

### Variables

Variables replacement can be applied to most entries.  It will use the <<< >>> notation.  

Variables can be loaded from a configuration file and/or from Extract and Capture directivess.

```
Meta:
  Config: variables.config
Workflow:
  - Ginfra:
      - Detach: true
      - Cmd: manage local start config:// --clear
      - Output:
        - Until: 4000
        - Source: log/ginfra.log
        - Scan: <<<expected.output>>>
```

It the above example, variables will be loaded from a file named variables.config.
If the variable expected.output exists, it will replace the Scan.

The file is a list of name/value pairs separated by =, as seen below:

```
variable1=value1
expected.output=Server starts
```

Basic variable replacement occurs with the variable name between <<< >>>.  Additional features
are also available.

### Variable replacement - environment variable,

The '^' will get the named environment variable.  If the variable doesn't exist, it will produce an empty string.
```
<<<^PATH>>>

```

The above example would produce the value of the $PATH (%PATH% on windows) environment  variable.

#### Variable replacement - configuration path.

The '$' directive allows you to query the loaded configuration provider.  Configuration is hierarchical and 
can be represented as json, as in [this real example](../design/config/config.json).
```
<<<$deployments.default.services.server_otelcol.service.config.service_class>>>

```

The above example would retrieve the value 'server_otelcol', if the configuration looked like the 
example linked above.  The nodes in the hierarchical tree are separated by periods (.).

#### Variable replacement - service discovery

The '=' directive will perform service discovery.

```
<<<=my_deployment_name.my_service_name>>>
```

With the deployment name and service name separated by a period (.).  Additionally, you can ask for the deployment
name to be supplied the following way:

```
<<<=%deployment.my_service_name>>>
```
 Using the % character prefixing the word 'deployment.'

#### Variable replacement - service class discovery

The '*' directive will perform service class discovery.

```
<<<*my_deployment_name.my_service_class_name>>>

```
The difference is you are asking by service class rather than service name.  Also you can give a wildcard for the first
segment of the service class like this: ```<<<*my_deployment_name.*/test>>>``` in which case you will get the first matching
service class for the deployment.  Note that the first part is usually the service name, but it can be something else.

You can use the % character the same as in service name discovery.

```
<<<*%deployment.my_service_class_name>>>
```

Using the % character prefixing the word 'deployment'.

You can see [examples of these in the gservices demos](https://gitlab.com/ginfra/gservices/-/blob/main/integration/demo/kube/demo.yaml).

### Variable replacement - management auth token

The '+' directive will fetch the management authorization token for a service.

```
<<<+my_deployment_name.my_service_name>>>

```

With the deployment name and service name separated by a period (.).  Additionally, you can ask for the deployment
name to be supplied the following way:

```
<<<+%deployment.my_service_name>>>
```
Using the % character prefixing the word 'deployment'.

### Variable replacement - service port

The '&' directive will fetch the service port for a service.

```
<<<&my_deployment_name.my_service_name>>>

```

With the deployment name and service name separated by a period (.).  Additionally, you can ask for the deployment
name to be supplied the following way:

```
<<<&%deployment.my_service_name>>>
```
Using the % character prefixing the word 'deployment'.




