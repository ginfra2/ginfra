//go:build unit_test

/*
Package config
Copyright (c) 2023 Erich Gatejen
[LICENSE]

Unit testing for config package.

[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt
*/
package config

import (
	"github.com/stretchr/testify/assert"
	testutil "gitlab.com/ginfra/ginfra/common/test"
	"reflect"
	"testing"
)

// ===============================================================================================================
// =

var (
	testConfigSimpleData = `{
	  "services": {
		"giuser": {
		  "default": {
			"Config": {
			  "image": "...",
			  "user_provider": "file://test/userfile.json",
			  "port": 8080,
              "sport": "8080",
	 	      "specifics": {
				"user_provider": "file://test/userfile.json"
			  }
			},
			"instances": {
			  "0": {
				"type": "local"
			  }
			}
		  },
		  "second": {
			"Config": {
			  "image": "...",
			  "user_provider": "file://test/userfile.json",
			  "port": 8080,
              "sport": "8080",
	 	      "specifics": {
				"user_provider": "file://test/userfile.json"
			  }
			},
			"instances": {
			  "0": {
				"type": "local"
			  }
			}
		  }
		}
	  }
	}`
)

func testAssertValueString(t *testing.T, c *Config, name string, value any) {
	v, err := c.GetValue(name).GetValueString()
	testutil.TestAssertValue(t, err, value, v)
}

func testAssertValueInt64(t *testing.T, c *Config, name string, value any) {
	v, err := c.GetValue(name).GetValueInt64()
	testutil.TestAssertValue(t, err, value, v)
}

func testConfig(t *testing.T) {
	var (
		cpf configProviderFile
		err error
	)
	cpf.savable = 0
	cpf.memoryOnly = true
	cpf.config, err = unmarshalGetConfigProviderFile([]byte(testConfigSimpleData))
	if err != nil {
		t.Error(err)
		return
	}

	root := cpf.FetchConfig(CR_DEPLOYMENTS)

	root.PutValue("number1", "val_number1")
	root.GetConfigPath("number1")
	if err != nil {
		t.Error(err)
		return
	}
	testAssertValueString(t, root.GetConfigPath("number1"), "number1", "val_number1")

	c2 := NewConfig(nil, nil)
	c2.PutValue("c2_1", "val_c2_1")
	c2.PutValue("c2_2", "val_c2_2")
	_ = root.PutConfigPath(c2, "giuser", "default")
	testAssertValueString(t, root.GetConfigPath("giuser", "default"), "c2_1", "val_c2_1")
	testAssertValueString(t, root.GetConfigPath("giuser", "default"), "c2_2", "val_c2_2")

	// Remove
	root.RemoveConfigPath("second", "Config")
	c3 := cpf.FetchConfig("services", "second", "Config")
	assert.NotEqual(t, nil, c3.Err, "Removed config should be gone.")
}

func testFileProvider(t *testing.T) {
	var (
		cpf configProviderFile
		err error
	)
	cpf.savable = 0
	cpf.memoryOnly = true
	cpf.config, err = unmarshalGetConfigProviderFile([]byte(testConfigSimpleData))
	if err != nil {
		t.Error(err)
		return
	}

	c1 := cpf.FetchConfig("services", "giuser", "default", "Config")
	testAssertValueInt64(t, c1, "port", int64(8080))
	testAssertValueInt64(t, c1, "sport", int64(8080))
	testAssertValueString(t, c1.GetConfig("specifics"), "user_provider", "file://test/userfile.json")
}

func TestConfigSimple(t *testing.T) {
	testConfig(t)
	testFileProvider(t)
}

type TestFakeStruct struct {
	bork int
}

func TestPutConfigPath(t *testing.T) {
	tests := []struct {
		name    string
		create  *Config
		input   *Config
		path    []string
		want    *Config
		wanterr error
	}{
		{
			name:    "root_modify",
			create:  &Config{Config: map[string]interface{}{"a_key": "a_value"}},
			input:   &Config{Config: map[string]interface{}{"new_key": "new_value"}},
			path:    []string{},
			want:    &Config{Config: map[string]interface{}{"a_key": "a_value", "new_key": "new_value"}},
			wanterr: nil,
		},
		{
			name:    "nested_modify",
			create:  &Config{Config: map[string]interface{}{"a_key": map[string]interface{}{"b_key": "b_value"}}},
			input:   &Config{Config: map[string]interface{}{"new_key": "new_value"}},
			path:    []string{"a_key"},
			want:    &Config{Config: map[string]interface{}{"a_key": map[string]interface{}{"b_key": "b_value", "new_key": "new_value"}}},
			wanterr: nil,
		},
		{
			name:    "nested_add",
			create:  &Config{Config: map[string]interface{}{"a_key": map[string]interface{}{"b_key": "b_value"}}},
			input:   &Config{Config: map[string]interface{}{"new_key": "new_value"}},
			path:    []string{"non_exist"},
			want:    &Config{Config: map[string]interface{}{"a_key": map[string]interface{}{"b_key": "b_value"}, "non_exist": map[string]interface{}{"new_key": "new_value"}}},
			wanterr: nil,
		},
		{
			name:    "path_length_zero",
			create:  &Config{Config: map[string]interface{}{"a_key": map[string]interface{}{"b_key": "b_value"}}},
			input:   &Config{Config: map[string]interface{}{"new_key": "new_value"}},
			path:    []string{""},
			want:    &Config{Config: map[string]interface{}{"a_key": map[string]interface{}{"b_key": "b_value"}, "new_key": "new_value"}},
			wanterr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.create.PutConfigPath(tt.input, tt.path...)
			if !reflect.DeepEqual(tt.create, tt.want) || !reflect.DeepEqual(err, tt.wanterr) {
				t.Fatal("Mismatch, Path: ", tt.path, "Got Config: ", tt.create, "Err: ", err, "Expected Config: ", tt.want, "Err: ", tt.wanterr)
			}
		})
	}
}

func TestConfigCases(t *testing.T) {

	t.Run("Case Config", func(t *testing.T) {
		var c Config
		c.Config = map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": "level2",
			},
		}

		r := c.GetConfigPath("level1", "level2", "level3")
		assert.Error(t, r.Err)

		err := c.PutConfigPath(&c, "level1", "level2", "level3")
		assert.Error(t, err)

		r = c.RemoveConfigPath("level1", "level2", "level3")
		assert.Error(t, r.Err)

		v := c.GetValue("level1")
		assert.Error(t, v.err)

		r = c.GetConfigPath("ply1")
		v = r.GetValue("")

	})

	t.Run("Case Config Value", func(t *testing.T) {
		var c ConfigValue

		c.value = map[string]interface{}{"bork": "borkity"}
		_, err := c.GetValueString()
		assert.Error(t, err)

		_, err = c.GetValueInt64()
		assert.Error(t, err)

		c.value = byte(1)
		_, err = c.GetValueString()
		assert.Error(t, err)

		_, err = c.GetValueInt64()
		assert.NoError(t, err) // It would appear that byte is an alias for uint8.

		c.value = TestFakeStruct{}
		v := c.GetValueInt64Noerr()
		assert.Equal(t, int64(0), v, "It should be 0 since it masks an error.")

		c.value = uint8(55)
		v, err = c.GetValueInt64()
		assert.NoError(t, err)
		assert.Equal(t, int64(55), v)

		c.value = true
		v, err = c.GetValueInt64()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), v)

		c.value = false
		v, err = c.GetValueInt64()
		assert.NoError(t, err)
		assert.Equal(t, int64(0), v)

	})

}
