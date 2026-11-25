//go:build functional_test

/*
Package ginterfaces
Copyright (c) 2023 Erich Gatejen
[LICENSE] https://www.apache.org/licenses/LICENSE-2.0.txt

Unit tests for openapi merge.
*/

package ginterfaces

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/ginfra/ginfra/base"
	testutil "gitlab.com/ginfra/ginfra/common/test"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenApiMergeLoadSpecs(t *testing.T) {

	td := filepath.Join(base.GetGinfraHome(), "app", "ginterface", "test", "mergetest")

	t.Run("Clean merge with resolved", func(t *testing.T) {
		d, _, err := OpenApiLoadSpecs(td, "gib2b", false, []string{"service_gib2b_1.yaml"})
		if err != nil {
			t.Error(err)
			return
		}
		//testutil.TestCheckMapValue(t, false, d, "gib2b_service", "info", "title")  this could be out of order
		testutil.TestCheckMapValue(t, false, d, "MIT License", "info", "license", "name")

		testutil.TestCheckMapValue(t, false, d, "Get specifics", "paths", "/ginfra/specifics", "get",
			"description")
		testutil.TestCheckMapValue(t, false, d, "Not found", "paths", "/gib2b/biz/{bizname}",
			"get", "responses", "404", "description")

		testutil.TestCheckMapValue(t, false, d, "string", "components", "schemas", "giusername", "type")
		testutil.TestCheckMapValue(t, false, d, "([A-Za-z0123456789+/=])+", "components",
			"schemas", "base64", "pattern")

	})

	t.Run("Nested resolve", func(t *testing.T) {
		d, _, err := OpenApiLoadSpecs(td, "gib2b", false, []string{"service_2_1.yaml"})
		if err != nil {
			t.Error(err)
			return
		}
		testutil.TestCheckMapValue(t, false, d, "object", "components", "schemas", "namevalue", "type")
		testutil.TestCheckMapValue(t, false, d, "string", "components", "schemas", "Specifics", "type")
		testutil.TestCheckMapValue(t, false, d, "string", "components", "schemas", "base64", "type")
	})

	t.Run("Empty list", func(t *testing.T) {
		d, _, err := OpenApiLoadSpecs(td, "gib2b", false, []string{})
		assert.NotNil(t, d)
		assert.NoError(t, err)
	})

}

func TestOpenApiMergeCases(t *testing.T) {
	p := filepath.Join(base.GetGinfraHome(), "app", "ginterface", "test", "mergetest")

	t.Run("openapiFindReadYamlFile spec in cwd", func(t *testing.T) {
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		defer func() {
			if err = os.Chdir(cwd); err != nil {
				panic(err)
			}
		}()
		if err = os.Chdir(p); err != nil {
			panic(err)
		}

		_, err = openapiFindReadYamlFile("/", "service_gib2b_1.yaml")
		assert.NoError(t, err)
	})

	t.Run("openapiFindReadYamlFile spec in ginfra", func(t *testing.T) {
		_, err := openapiFindReadYamlFile("/", "ginfra.yaml")
		assert.NoError(t, err)
	})

	t.Run("openapiFindReadYamlFile spec does not exist", func(t *testing.T) {
		_, err := openapiFindReadYamlFile("/", "aseofkawokfoqkfewokfwov.yaml")
		assert.ErrorContains(t, err, "Does not exist")
	})

	t.Run("openapiFindReadYamlFile bad yaml", func(t *testing.T) {
		_, err := openapiFindReadYamlFile(p, "bad.yaml")
		assert.ErrorContains(t, err, "Bad yaml")
	})

	t.Run("openApiSeekRef", func(t *testing.T) {
		r := map[string]struct{}{}
		openApiSeekRef(map[string]interface{}{"$ref": "common.yaml#/thing/thang"}, r)
		assert.Less(t, 0, len(r))
		if _, ok := r["common.yaml"]; !ok {
			t.Error("Resolve missing.")
		}
	})

	t.Run("openapiFindReadYamlFile bad yaml", func(t *testing.T) {
		_, err := openapiFindReadYamlFile(p, "bad.yaml")
		assert.ErrorContains(t, err, "Bad yaml")
	})

	t.Run("openapiMergeArraySection bad section", func(t *testing.T) {
		_, err := openapiMergeArraySection("x", map[string]interface{}{}, []interface{}{})
		assert.ErrorContains(t, err, "Source section is not an array")
	})

	t.Run("openapiGetOrMakeObject not object", func(t *testing.T) {
		_, err := openapiGetOrMakeObject("name", map[string]interface{}{"name": "value"})
		assert.ErrorContains(t, err, "Section is not an object")
	})

	t.Run("openapiGetOrMakeArray not array", func(t *testing.T) {
		_, err := openapiGetOrMakeArray("name", map[string]interface{}{"name": "value"})
		assert.ErrorContains(t, err, "Section is not an array")
	})

}

func TestOpenapiMergeObjectSection(t *testing.T) {
	var overwrite []string
	tests := []struct {
		name          string
		in            interface{}
		target        map[string]interface{}
		wantErr       bool
		wantOverwrite bool
	}{
		{
			name:    "Case: Empty map",
			in:      make(map[string]interface{}),
			target:  make(map[string]interface{}),
			wantErr: false,
		},
		{
			name:    "Case: Source section not an object",
			in:      "wrong value",
			target:  make(map[string]interface{}),
			wantErr: true,
		},
		{
			name: "Case: Overwriting existing value",
			in: map[string]interface{}{
				"test_key": "test_value",
			},
			target: map[string]interface{}{
				"test_key": "old_value",
			},
			wantErr:       false,
			wantOverwrite: true,
		},
		{
			name: "Case: No Overwriting",
			in: map[string]interface{}{
				"test_key": "test_value",
			},
			target: map[string]interface{}{
				"another_key": "any_value",
			},
			wantErr: false,
		},
		{
			name: "Case: Merging two maps with distinct keys",
			in: map[string]interface{}{
				"test_key": "test_value",
			},
			target: map[string]interface{}{
				"another_key": "any_value",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := openapiMergeObjectSection(tt.name, tt.in, tt.target, overwrite)
			if (err != nil) != tt.wantErr {
				t.Errorf("openapiMergeObjectSection() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if _, ok := tt.in.(map[string]interface{}); ok {
				if _, exist := tt.target["test_key"]; exist && len(overwrite) > 0 {
					if overwrite[0] != "Overwrite: test_key in "+tt.name {
						t.Errorf("openapiMergeObjectSection() overwrite = %v, want %v", overwrite[0], "Overwrite: test_key in "+tt.name)
					}
				}
			} else if err == nil {
				t.Errorf("openapiMergeObjectSection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOpenapiMergeSpecs(t *testing.T) {
	t.Run("info get object error", func(t *testing.T) {
		_, err := openapiMergeSpecs("x", map[string]interface{}{"info": "xxx"},
			map[string]interface{}{"info": "xxx"}, []string{})
		assert.ErrorContains(t, err, "Section is not an object.")
	})
	t.Run("info get merge error", func(t *testing.T) {
		_, err := openapiMergeSpecs("x", map[string]interface{}{"info": "xxx"},
			map[string]interface{}{}, []string{})
		assert.ErrorContains(t, err, "section is not an object.")
	})
	t.Run("security get object error", func(t *testing.T) {
		_, err := openapiMergeSpecs("x", map[string]interface{}{"security": "xxx"},
			map[string]interface{}{"info": "xxx"}, []string{})
		assert.ErrorContains(t, err, "section is not an array.")
	})
	t.Run("security get merge error", func(t *testing.T) {
		_, err := openapiMergeSpecs("x", map[string]interface{}{"security": "xxx"},
			map[string]interface{}{}, []string{})
		assert.ErrorContains(t, err, "section is not an array.")
	})
	t.Run("components error", func(t *testing.T) {
		_, err := openapiMergeSpecs("x", map[string]interface{}{"components": "xxx"},
			map[string]interface{}{}, []string{})
		assert.ErrorContains(t, err, "not an object")
	})
	t.Run("components get object error", func(t *testing.T) {
		_, err := openapiMergeSpecs("x",
			map[string]interface{}{"components": map[string]interface{}{"info": "xxx"}},
			map[string]interface{}{"components": "xxx"}, []string{})
		assert.ErrorContains(t, err, "Section is not an object")
	})
	t.Run("components get object error", func(t *testing.T) {
		_, err := openapiMergeSpecs("x",
			map[string]interface{}{"components": map[string]interface{}{"info": "xxx"}},
			map[string]interface{}{}, []string{})
		assert.ErrorContains(t, err, "Section in Components is not an Object.")
	})
	t.Run("components get merge object error", func(t *testing.T) {
		_, err := openapiMergeSpecs("x",
			map[string]interface{}{"components": map[string]interface{}{"info": map[string]interface{}{"thing": "xxx"}}},
			map[string]interface{}{"components": map[string]interface{}{"info": "sdf"}},
			[]string{})
		assert.ErrorContains(t, err, "Section is not an object.")
	})
}
