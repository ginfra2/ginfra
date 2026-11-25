## GINFA COMMAND : TOOL

### TOOL INFO

ginfra tool info

This will print information about the ginfra app.  It isn't particularly compelling at this time, as it was
added to enable testing.

### TOOL SELENIUM

A tool for interacting to gservices/service_selenium.  It can be used to drive
traffic and run certain tests using (selenium)[https://www.selenium.dev/].

#### TOOL SELENIUM RUN

ginfra tool selenium run [host] [target] [path to snippet]

[host] is the running gservices/service_selenium service instance.  

[target] is the URL to the website to be driven by selenium.

[path to snippet] is a snippet of java code that drives selenium.  Below is
an example of a snippet.

```
    driver.findElement(By.cssSelector(".inputusername")).click();
    driver.findElement(By.cssSelector(".inputusername")).sendKeys("bob"); 
    for (int i = 0; i < 1000; i++) {
        driver.findElement(By.cssSelector(".doit")).click();
    }
```

The service will handle everything necessary to get a working driver and to clean up
after it this runs.  It does not have to exit gracefully and can be interrupted by another command.

The following is an example of the command.
```
ginfra tool selenium run 192.168.1.220:8904 http://192.168.1.220:8080/ /build/gservices/integration/workbench/otel/host/e2e.snp
```

#### TOOL SELENIUM STOP

ginfra tool selenium stop [host] [id] 

[host] is the running gservices/service_selenium service instance.  

[id] is not currently used and can be any text.

This will stop any running snippet.

The following is an example of the command.
```
ginfra tool selenium stop 192.168.1.220:8904 x
```

#### TOOL SELENIUM STATUS

ginfra tool selenium status [host]

[host] is the running gservices/service_selenium service instance.  

This will report the status of the service and if a snippet is running or not.  If there
was an error in the last snippet, it will be reported.

The following is an example of the command.
```
ginfra tool selenium status 192.168.1.220:8904
```

