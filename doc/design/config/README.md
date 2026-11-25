### CLUSTER/DEPLOYMENT CONFIG

The giconfig service in the ginfra/gservices repo owns the configuration so any changes made to the 
definition will have a direct impact to that service.   Additionally, the config provider that is 
part of ginfra/common in this repo could be affected.  

The config.json shows a current example of the configuration with one service registered.

### SERVICE CONFIG

This is the json that is passed to a service when started.  This is the mechanism to pass config
from the management tools (like ginfra).  [service_config.json](service_config.json) is an example.


