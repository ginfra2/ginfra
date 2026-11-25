/*
Package cmd
Copyright (c) 2023 Erich Gatejen
[LICENSE]: https://www.apache.org/licenses/LICENSE-2.0.txt

Specifics tests.
*/
package shared

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	"gitlab.com/ginfra/ginfra/common"
	"path/filepath"
	"testing"
)

var testSpecifics = `{
			   "meta": {
				  "version": 1,
				  "name": "configservice",
				  "type": "apiecho",
				  "module_name": "gitlab.com/ginfra/gservices/service_giconfig",
				  "serviceclass": "configservice",
				  "invocationcmd": ""
			   },
			   "exposed": {
				  "4317":{
					 "name": "Otel Collector receiver grpc",
					 "serviceclass":  {
						"name": "collector",
						"url": "grpc://<<<host>>>:<<<port>>>/"
					 }
				  },
				  "4318":{
					 "name": "Otel Collector receiver http"
				  }
			   }
			}`

var testSpecificsBadReplacement = `{
               "meta": {
				  "version": 1,
				  "name": "configservice",
				  "type": "apiecho",
				  "module_name": "gitlab.com/ginfra/gservices/service_giconfig",
				  "serviceclass": "configservice",
				  "invocationcmd": ""
			   }
			   "exposed": {
				  "4317":{
					 "name": "Otel Collector receiver grpc",
					 "serviceclass":  {
						"name": "collector",
						"url": "grpc://<<<host<<<port/"
					 }
				  },
				  "4318":{
					 "name": "Otel Collector receiver http"
				  }
			   }
			}`

var testSpecificsMissingType = `{
               "meta": {
				  "version": 1,
				  "name": "configservice",
				  "type": "",
				  "module_name": "gitlab.com/ginfra/gservices/service_giconfig",
				  "serviceclass": "configservice",
				  "invocationcmd": ""
			   }}`

var testSpecificsMissingServiceClass = `{
               "meta": {
				  "version": 1,
				  "name": "configservice",
				  "type": "apiecho",
				  "module_name": "gitlab.com/ginfra/gservices/service_giconfig",
				  "serviceclass": "",
				  "invocationcmd": ""
			   }}`

func init() {
	common.SetupTest()
}

func TestSpecificsLoad(t *testing.T) {
	testCases := []struct {
		name        string
		inputString string
		inputPath   string
		errString   string
	}{
		{
			name:        "ValidJSON",
			inputString: testSpecifics,
			inputPath:   "valid.json",
			errString:   "",
		},
		{
			name:        "InvalidJSON",
			inputString: `invalid_json`,
			inputPath:   "invalid.json",
			errString:   "Could not load",
		},
		{
			name:        "EmptyJSON",
			inputString: `{}`,
			inputPath:   "empty.json",
			errString:   "",
		},
		{
			name:        "Bad replacement",
			inputString: testSpecificsBadReplacement,
			inputPath:   "testSpecificsBadReplacement",
			errString:   "replacement problem",
		},
		{
			name:        "missing type",
			inputString: testSpecificsMissingType,
			inputPath:   "testSpecificsMissingType",
			errString:   "does not contain a type",
		},
		{
			name:        "missing serviceclass",
			inputString: testSpecificsMissingServiceClass,
			inputPath:   "testSpecificsMissingServiceClass",
			errString:   "does not contain a serviceclass",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmp, err := SpecificsLoad(tc.inputString, tc.inputPath, map[string]interface{}{})

			if tc.errString != "" {
				assert.ErrorContains(t, err, tc.errString)
			}

			if tmp != nil {
				v, ok := tmp.Exposed["4317"]
				if ok {
					assert.Equal(t, "Otel Collector receiver grpc", v.Name)
					assert.Equal(t, "collector", v.ServiceClass.Name)
					assert.Equal(t, "grpc://<<<host>>>:<<<port>>>/", v.ServiceClass.Url)
				}
			}

		})
	}
}

func TestSpecificsFile(t *testing.T) {

	s, err := SpecificsLoad(testSpecifics, "blah", map[string]interface{}{})
	if err != nil {
		t.Error(err)
		return
	}

	f := filepath.Join(base.GetGinfraTmpDir(), base.GetRandomFileName(".json"))
	err = s.WriteFile(f)
	if err != nil {
		t.Error(err)
		return
	}

	h, err := SpecificsLoadFile(f, map[string]interface{}{})
	if err != nil || s == nil {
		t.Error(err)
		return
	}

	v, ok := h.Exposed["4317"]
	if ok {
		assert.Equal(t, "Otel Collector receiver grpc", v.Name)
	} else {
		t.Error("4317 not found.")
	}

}

func TestSpecificsCases(t *testing.T) {

	t.Run("Bad file", func(t *testing.T) {
		var val map[string]interface{}
		_, err := SpecificsLoadFile("sdfasdfas", val)
		assert.ErrorContains(t, err, "Specifics file does not exist")
	})

	t.Run("Bad write file", func(t *testing.T) {
		var s Specifics
		err := s.WriteFile("../../../../../../../../../../../../../../../../../../../../../../../../..")
		assert.ErrorContains(t, err, "Could not write specifics file")
	})

	t.Run("Cannot read file", func(t *testing.T) {
		var val map[string]interface{}
		_, err := SpecificsLoadFile("/etc", val)
		assert.ErrorContains(t, err, "Could not read")
	})

}
