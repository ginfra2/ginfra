/*
Package config
Copyright (c) 2023, 2024 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Local, in-memory configuration management.  A configuration provider (which is also available) manages getting these
in-memory instances into and out of another storage mechanism.  It is largely interface{} based because the data
itself has a strong affinity to JSON.  For ephemeral configuration, you can instantiate a Config struct using
NewConfig and get/put from it as long as you want.   Likewise there is a loose connection between Config class
and the providers, where you dont have to store or retrieve stored config until you are ready.

Configuration is hierarchical and the following is a value configuration set.  Authorization is handled by the
providers (if done at all), but because the configuration has to occur before joining any OAUTH scheme, it has to
have it's own AUTH scheme.  You see some hints of that in the example below and in the code.

Technically, this can be used with any configuration layout, but below is an example of how the Ginfra app uses it.

```

	{
	    "auth": "ACYTYHVCJTDN",
	    "deployments": {
	        "default": {
				"values" : {
					"name1": "value1",
					"name2": "value2",
				}
	            "discovery": {
	                "portals": {
	                    "*": "node1"
	                },
	                "services": {
	                    "giuser": "http://192.168.144.3:8900"
	                }
	            },
	            "services": {
	                "giuser": {
	                    "management": {
	                        "auth": "XIBPMIRPRTAG",
	                        "instance": {
	                            "fqdn": "192.168.144.3",
	                            "id": "ca01bd12fbc5de76e3e4a023b6b8897db6711da085f5167933d92f65e354ce15",
	                            "ip": "192.168.144.3",
	                            "type": "host"
	                        }
	                    },
	                    "service": {
	                        "auth": "XIBPMIRPRTAG",
	                        "config": {
	                            "config_provider": "http://192.168.144.2:5111/",
	                            "image": "gservices/giuser",
	                            "port": 8900,
	                            "service_class": "giuser",
	                            "specifics": {
	                                "user_provider": "file:///service/test/userfile.json"
	                            }
	                        },
	                        "service_auth": "II3SNFNGN2TOLSF1O462"
	                    }
	                }
	            }
	        }
	    },
	    "live": {
	        "time": "2024-05-01T15:32:52.021759154Z"
	    }
	}

```
*/
package config

import (
	"fmt"
	"gitlab.com/ginfra/ginfra/base"
	"reflect"
	"strconv"
	"strings"
)

// #####################################################################################################################
// # GENERAL CONFIG

const AUTH_TOKEN_LENGTH = 12
const AUTH_ADMIN_TOKEN_DEFAULT = "AAAAAAAAAAAA"
const CONFIG_NIL_AUTH = AUTH_ADMIN_TOKEN_DEFAULT
const AUTH_USER_TOKEN_DEFAULT = "XXXXXXXXXXXX"
const CONFIG_RO_FLAG = "X"

const RESERVED_CONFIG_SERVICE_PORT = 5111
const DEFAULT_SERVICE_PORT = 8900

// #####################################################################################################################
// # CONFIGURATION LAYOUTS - Probably belongs in another file.  TBD.

var DeploymentCreateJSON = `
{
	"services": {
	},
	"discovery": {
		"services": {
		},
		"portals": {
			"*": "localhost"
		}
	}
}`

var ServiceRegisterRequestJSON string = `
{
	  "service": {
		  "config": {
			  "image": "<<< image >>>",
			  "port": <<< port >>>,
			  "config_provider": "<<< config_provider >>>",
              "service_class": "<<< service_class >>>",
			  "specifics": {
			  }
		  },
		  "auth": "<<< auth >>>",
          "service_auth": "<<< service_auth >>>"
	  }
}`

// ServiceHostInstanceRequestJSON represents the JSON string for registering a service host instance.  There will only
// be one instance with instance type 'host' so this will never change once set.
var ServiceHostInstanceRequestJSON string = `
{
	  "management": {
		  "instance": {
		  }
	  }
}`

// #####################################################################################################################
// # CONFIGURATION PROVIDER PUBLIC API

var (
	ConfigNotFound = "Config not found."
	//ConfigPathNotRequestedType   = "Config path not requested type."
	ConfigPathIsConfigNotValue   = "Config path points to more Config, not a value."
	ConfigPathNodeValueNotConfig = "Config pathnode is a value not a Config."
	ConfigUnsupportedType        = "Config path contains an unsupported type."
)

// Config struct.  It carries Err around, so we can chain calls for convenience.
type Config struct {
	Config map[string]interface{}
	Err    error
}

// ConfigValue struct.  It also carries err, so we can chain calls.
type ConfigValue struct {
	value any
	name  string
	err   error
}

// NewConfig initializes a new Config instance with the provided map and error values.
func NewConfig(c map[string]interface{}, e error) *Config {
	var config Config
	if c == nil {
		config.Config = make(map[string]interface{})
	} else {
		config.Config = c
	}
	config.Err = e
	return &config
}

// GetConfig retrieves a sub-configuration by the specified name and returns it as a new Config instance.
// If the name doesn't exist or an error occurs, the returned Config contains the corresponding error.
func (c *Config) GetConfig(name string) *Config {
	var (
		r Config
	)
	if item, ok := c.Config[name]; c.Err == nil && ok {

		switch item.(type) {

		case map[string]interface{}:
			r = Config{
				Config: item.(map[string]interface{}),
				Err:    c.Err,
			}
			break

		default:
			m := make(map[string]interface{})
			m[name] = item
			r = Config{
				Config: m,
				Err:    c.Err,
			}
			break
		}
	} else {
		r = Config{
			Config: nil,
			Err:    base.NewGinfraErrorAFlags(ConfigNotFound, base.ERRFLAG_NOT_FOUND, base.LM_CONFIG_NAME, name),
		}
	}

	return &r
}

// GetConfigPath retrieves a nested configuration by traversing the specified path and returns it as a new Config instance.
// If an intermediate path node is not a valid configuration, an error is set in the returned Config instance.
func (c *Config) GetConfigPath(path ...string) *Config {
	var p []string
	if len(path) == 1 {
		p = strings.Split(path[0], ".")
	} else {
		p = path
	}

	r := NewConfig(c.Config, c.Err)
	for i, item := range p {
		if i < len(path)-1 {
			if n, ok := r.Config[item]; ok {
				if rn, ok := n.(map[string]interface{}); ok {
					r.Config = rn
				} else {
					r.Err = base.NewGinfraErrorAFlags(ConfigPathNodeValueNotConfig, base.ERRFLAG_NOT_FOUND,
						base.LM_CONFIG_NAME, item)
					break
				}
			}
		} else {
			r = r.GetConfig(item)
		}
	}

	return r
}

// PutConfigPath inserts the given configuration into the specified path in the current Config instance, creating
// missing nodes.
func (c *Config) PutConfigPath(config *Config, path ...string) error {
	var (
		p   []string
		r   *Config
		err error
	)
	if len(path) == 1 {
		p = strings.Split(path[0], ".")
	} else {
		p = path
	}

	// Trivial case
	if len(p) < 1 || (len(p) == 1 && p[0] == "") {
		for n, v := range config.Config {
			c.PutValue(n, v)
		}

	} else {
		r = NewConfig(c.Config, c.Err)
		for _, item := range p {
			if n, ok := r.Config[item]; ok {
				if rn, ok := n.(map[string]interface{}); ok {
					r.Config = rn
				} else {
					r.Err = base.NewGinfraErrorAFlags(ConfigPathNodeValueNotConfig, base.ERRFLAG_NOT_FOUND,
						base.LM_CONFIG_NAME, item)
					break
				}
			} else {
				r.Config[item] = make(map[string]interface{})
				r.Config = r.Config[item].(map[string]interface{})
			}
		}
		if r.Err == nil {
			for n, v := range config.Config {
				r.Config[n] = v
			}
		} else {
			err = r.Err
		}
	}
	return err
}

// RemoveConfigPath deletes a configuration entry at the specified path and sets an error if the entry does not exist.
func (c *Config) RemoveConfigPath(path ...string) *Config {
	var (
		cn   *Config
		last string
	)

	if len(path) < 2 {
		cn = c
		last = path[0]
	} else {
		cn = c.GetConfigPath(path[:len(path)-1]...)
		last = path[len(path)-1]
	}

	if cn.Err == nil {
		if _, ok := cn.Config[last]; ok {
			delete(cn.Config, last)
		} else {
			cn.Err = base.NewGinfraErrorAFlags(ConfigNotFound, base.ERRFLAG_NOT_FOUND, base.LM_CONFIG_NAME, last)
		}
	}

	return cn
}

// PutValue adds or updates a key-value pair in the Config. If the Config map is nil, it initializes it.
func (c *Config) PutValue(name string, value any) {
	if c.Config == nil {
		c.Config = make(map[string]interface{})
	}
	c.Config[name] = value
}

// GetValue retrieves the value associated with the specified key from the Config instance and returns it as a ConfigValue.
func (c *Config) GetValue(name string) *ConfigValue {
	var (
		r = ConfigValue{nil, name, c.Err}
	)

	if item, ok := c.Config[name]; r.err == nil && ok {

		switch ItemType := item.(type) {

		case map[string]interface{}:
			r.err = base.NewGinfraErrorA(ConfigPathIsConfigNotValue, base.LM_CONFIG_NAME, name)
			break

		default:
			_ = ItemType
			r.value = item
			break
		}
	} else {
		r.err = base.NewGinfraErrorAFlags(ConfigNotFound, base.ERRFLAG_NOT_FOUND, base.LM_CONFIG_NAME, name)
	}

	return &r
}

// GetValueError returns the error associated with the ConfigValue, if any.
func (c *ConfigValue) GetValueError() error {
	return c.err
}

// GetValueStringNoerr returns the string value of the ConfigValue without handling any errors that may occur.
func (c *ConfigValue) GetValueStringNoerr() string {
	v, _ := c.GetValueString()
	// TODO should I panic?
	return v
}

// GetValueString gets a value as a string.  The supported types match what most json libraries use.
func (c *ConfigValue) GetValueString() (string, error) {
	var (
		r   string
		err = c.err
	)
	if err == nil {

		switch ItemType := c.value.(type) {

		case map[string]interface{}:
			err = base.NewGinfraErrorA(ConfigPathIsConfigNotValue, base.LM_CONFIG_NAME, c.name)
			break

		case string:
			r = c.value.(string)
			break

		case bool, int, int32, int64, float64:
			r = fmt.Sprintf("%v", c.value)
			break

		default:
			err = base.NewGinfraErrorA(ConfigUnsupportedType, base.LM_CONFIG_NAME, c.name)
			fmt.Printf("ConfigUnsupportedType %v", ItemType)
			break
		}

	}
	return r, err
}

// GetValueInt64Noerr returns the int64 value of the ConfigValue, ignoring any errors that may occur.
func (c *ConfigValue) GetValueInt64Noerr() int64 {
	v, _ := c.GetValueInt64()
	// TODO should I panic?
	return v
}

// GetValueInt64 converts the value stored in ConfigValue to int64, returning an error if the conversion is not possible.
func (c *ConfigValue) GetValueInt64() (int64, error) {
	var (
		r   int64
		err error = c.err
	)
	if err == nil {

		switch ItemType := c.value.(type) {

		case map[string]interface{}:
			err = base.NewGinfraErrorA(ConfigPathIsConfigNotValue, base.LM_CONFIG_NAME, c.name)

		case string:
			r, err = strconv.ParseInt(c.value.(string), 10, 0)

		case float32, float64:
			// Cannot reliable convert float to int.  TODO deal with this.
			r = int64(reflect.ValueOf(c.value).Float())

		case bool:
			if reflect.ValueOf(c.value).Bool() {
				r = 1
			} else {
				r = 0
			}

		case uint, uint8, uint16, uint32:
			r = int64(reflect.ValueOf(c.value).Uint())

		case uint64:
			// Cannot reliable convert uint64 into int64.  TODO deal with this.
			r = int64(reflect.ValueOf(c.value).Uint())

		case int, int8, int16, int32, int64:
			r = reflect.ValueOf(c.value).Int()

		default:
			err = base.NewGinfraErrorA(ConfigUnsupportedType, base.LM_CONFIG_NAME, c.name)
			fmt.Printf("ConfigUnsupportedType %v", ItemType)

		}
	}

	return r, err
}
