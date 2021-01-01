package config_test

import (
	"fmt"
	"github/mlyahmed.io/nominee/infra"
	"github/mlyahmed.io/nominee/pkg/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type configurationExamples struct {
	empties  map[string]interface{}
	missing  []string
	defaults map[string]interface{}
	values   map[string]interface{}
}

const (
	environment = "Environment"
	file        = "File"
)

var (
	stringConfExamples = configurationExamples{
		empties:  map[string]interface{}{"STRING_CONF_1": "", "STRING_CONF_2": "", "STRING_CONF_3": ""},
		missing:  []string{"STRING_CONF_1", "STRING_CONF_2", "STRING_CONF_3"},
		defaults: map[string]interface{}{"STRING_CONF_1": "DEFAULT_1", "STRING_CONF_2": "DEFAULT_2", "STRING_CONF_3": "DEFAULT_3"},
		values:   map[string]interface{}{"STRING_CONF_1": "VALUE_1", "STRING_CONF_2": "VALUE_2", "STRING_CONF_3": "VALUE_3"},
	}

	intConfExamples = configurationExamples{
		empties:  map[string]interface{}{"INT_CONF_1": "", "INT_CONF_2": "", "INT_CONF_3": ""},
		missing:  []string{"INT_CONF_1", "INT_CONF_2", "INT_CONF_3"},
		defaults: map[string]interface{}{"INT_CONF_1": 3254, "INT_CONF_2": 2379, "INT_CONF_3": 80},
		values:   map[string]interface{}{"INT_CONF_1": 22, "INT_CONF_2": 8080, "INT_CONF_3": 443},
	}
)

func TestGetStringOrPanic_panics_when_the_key_is_defined_with_empty_string(t *testing.T) {
	t.Run(environment, func(t *testing.T) {
		t.Logf("Given the keys %s are loaded with empty values in the environment.", stringConfExamples.empties)
		{
			saveConfigurationsInTheEnvironment(stringConfExamples.empties, t)
			defer tearsDown(stringConfExamples.empties)
			for key, _ := range stringConfExamples.empties {
				thenGetStringOrPanicMustPanic(key, t)
			}
		}
	})

	t.Run(file, func(t *testing.T) {
		t.Logf("Given the keys %s are loaded with empty values in the file.", stringConfExamples.empties)
		{
			dir := saveConfigurationsInFile(stringConfExamples.empties, t)
			defer os.RemoveAll(dir)
			defer tearsDown(stringConfExamples.empties)
			for key, _ := range stringConfExamples.empties {
				thenGetStringOrPanicMustPanic(key, t)
			}
		}
	})
}

func TestGetStringOrPanic_panics_when_get_missing_string_value(t *testing.T) {
	t.Logf("Given these missing keys %s.", stringConfExamples.missing)
	{
		for _, key := range stringConfExamples.missing {
			thenGetStringOrPanicMustPanic(key, t)
		}
	}
}

func TestGetStringOrPanic_returns_value_when_the_key_is_defined(t *testing.T) {
	t.Run(environment, func(t *testing.T) {
		t.Logf("Given the searshed keys/values %s are loaded in the environment", stringConfExamples.values)
		{
			saveConfigurationsInTheEnvironment(stringConfExamples.values, t)
			defer tearsDown(stringConfExamples.values)
			for key, _ := range stringConfExamples.values {
				config.SetDefault(key, "THE DEFAULT VALUE THAT MUST BE IGNORED")
			}

			thenGetStringOrPanicMustReturnTheExpectedValue(t, stringConfExamples.values)
		}
	})

	t.Run(file, func(t *testing.T) {
		t.Logf("Given the searshed keys/values %s are loaded in the file.", stringConfExamples.values)
		{
			dir := saveConfigurationsInFile(stringConfExamples.values, t)
			defer os.RemoveAll(dir)
			defer tearsDown(stringConfExamples.values)
			for k, _ := range stringConfExamples.values {
				config.SetDefault(k, "THE DEFAULT VALUE THAT MUST BE IGNORED")
			}

			thenGetStringOrPanicMustReturnTheExpectedValue(t, stringConfExamples.values)
		}
	})

}

func TestGetStringOrPanic_returns_the_default_Value(t *testing.T) {
	defer tearsDown(stringConfExamples.defaults)
	t.Logf("Given the searshed keys/defaults %s are declared.", stringConfExamples.defaults)
	{
		for k, v := range stringConfExamples.defaults {
			config.SetDefault(k, v)
		}

		thenGetStringOrPanicMustReturnTheExpectedValue(t, stringConfExamples.defaults)
	}
}

func thenGetStringOrPanicMustPanic(key string, t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("\t%s\tFAIL: GetStringOrPanic(%s). Expected the program to panic. Actual not.", infra.Failed, key)
		} else {
			t.Logf("\t%s\tThen the program must panic.", infra.Succeed)
		}
	}()

	t.Logf("\tWhen get the value of the key %s", key)
	config.GetStringOrPanic(key)
}

func thenGetStringOrPanicMustReturnTheExpectedValue(t *testing.T, cases map[string]interface{}) {
	for k, expected := range cases {
		t.Logf("\tWhen Get the key %s value.", k)
		actualVal := config.GetStringOrPanic(k)
		if actualVal != expected {
			t.Fatalf("\t%s\tFAIL: GetStringOrPanic(%s) expected %s. Actual %s ", infra.Failed, k, expected, actualVal)
		}
		t.Logf("\t%s\tThen the value must be %s", infra.Succeed, expected)
	}
}

func TestGetString_returns_empty_string_when_the_key_is_not_defined(t *testing.T) {
	defer tearsDown(stringConfExamples.empties)

	t.Logf("Given the missing keys %s.", stringConfExamples.empties)
	{
		thenGetStringMustReturnTheExpectedValue(stringConfExamples.empties, t)
	}
}

func TestGetString_returns_the_default_value(t *testing.T) {
	defer tearsDown(stringConfExamples.defaults)

	t.Logf("Given the searshed keys/defaults %s are declared.", stringConfExamples.defaults)
	{
		for k, v := range stringConfExamples.defaults {
			config.SetDefault(k, v)
		}

		thenGetStringMustReturnTheExpectedValue(stringConfExamples.defaults, t)
	}
}

func TestGetString_returns_the_defined_value(t *testing.T) {
	t.Run("Environment", func(t *testing.T) {
		t.Logf("Given the searshed keys/values %s are loaded in the environment.", stringConfExamples.values)
		{
			saveConfigurationsInTheEnvironment(stringConfExamples.values, t)
			defer tearsDown(stringConfExamples.values)
			for key, _ := range stringConfExamples.values {
				config.SetDefault(key, "THE DEFAULT VALUE THAT MUST BE IGNORED ENV")
			}
			thenGetStringMustReturnTheExpectedValue(stringConfExamples.values, t)
		}
	})

	t.Run("File", func(t *testing.T) {
		t.Logf("Given the searshed keys/values %s are loaded in a configuration file.", stringConfExamples.values)
		{
			dir := saveConfigurationsInFile(stringConfExamples.values, t)
			defer os.RemoveAll(dir)
			defer tearsDown(stringConfExamples.values)
			for k, _ := range stringConfExamples.values {
				config.SetDefault(k, "THE DEFAULT VALUE THAT MUST BE IGNORED FILE")
			}

			thenGetStringMustReturnTheExpectedValue(stringConfExamples.values, t)
		}
	})
}

func thenGetStringMustReturnTheExpectedValue(cases map[string]interface{}, t *testing.T) {
	for k, expected := range cases {
		t.Logf("\tWhen Get the key %s value.", k)
		actualVal := config.GetString(k)
		if actualVal != expected {
			t.Fatalf("\t%s\tFAIL: GetString(%s) expected %s. Actual %s ", infra.Failed, k, expected, actualVal)
		}
		t.Logf("\t%s\tThen the value must be %s", infra.Succeed, expected)
	}
}

func TestGetIntOrPanic_panics_when_the_key_is_defined_with_empty_value(t *testing.T) {
	t.Run("Environment", func(t *testing.T) {
		t.Logf("Given the keys %s are loaded with empty values in the environment.", intConfExamples.empties)
		{
			saveConfigurationsInTheEnvironment(intConfExamples.empties, t)
			defer tearsDown(intConfExamples.empties)
			for key, _ := range intConfExamples.empties {
				thenGetIntOrPanicMustPanic(key, t)
			}
		}
	})

	t.Run("File", func(t *testing.T) {
		t.Logf("Given the keys %s are loaded with empty values in the file.", intConfExamples.empties)
		{
			dir := saveConfigurationsInFile(intConfExamples.empties, t)
			defer os.RemoveAll(dir)
			defer tearsDown(intConfExamples.empties)
			for key, _ := range intConfExamples.empties {
				thenGetIntOrPanicMustPanic(key, t)
			}
		}
	})
}

func TestGetIntOrPanic_panics_when_get_missing_string_value(t *testing.T) {
	t.Logf("Given these missing keys %s.", intConfExamples.missing)
	{
		for _, key := range intConfExamples.missing {
			thenGetIntOrPanicMustPanic(key, t)
		}
	}
}

func TestGetIntOrPanic_returns_value_when_the_key_is_defined(t *testing.T) {
	t.Run("Environment", func(t *testing.T) {
		t.Logf("Given the searshed keys/values %s are loaded in the environment", intConfExamples.values)
		{
			saveConfigurationsInTheEnvironment(intConfExamples.values, t)
			defer tearsDown(intConfExamples.values)
			for key, _ := range intConfExamples.values {
				config.SetDefault(key, "THE DEFAULT VALUE THAT MUST BE IGNORED")
			}

			thenGetIntOrPanicMustReturnTheExpectedValue(intConfExamples.values, t)
		}
	})

	t.Run("File", func(t *testing.T) {
		t.Logf("Given the searshed keys/values %s are loaded in the file.", intConfExamples.values)
		{
			dir := saveConfigurationsInFile(intConfExamples.values, t)
			defer os.RemoveAll(dir)
			defer tearsDown(intConfExamples.values)
			for k, _ := range intConfExamples.values {
				config.SetDefault(k, "THE DEFAULT VALUE THAT MUST BE IGNORED")
			}

			thenGetIntOrPanicMustReturnTheExpectedValue(intConfExamples.values, t)
		}
	})

}

func TestGetIntOrPanic_returns_the_default_value(t *testing.T) {
	defer tearsDown(intConfExamples.defaults)

	t.Logf("Given the searshed keys/defaults %s are declared.", intConfExamples.defaults)
	{
		for k, v := range intConfExamples.defaults {
			config.SetDefault(k, v)
		}

		thenGetIntOrPanicMustReturnTheExpectedValue(intConfExamples.defaults, t)
	}
}

func thenGetIntOrPanicMustPanic(key string, t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("\t%s\tFAIL: GetIntOrPanic(%s). Expected the program to panic. Actual not.", infra.Failed, key)
		} else {
			t.Logf("\t%s\tThen the program must panic.", infra.Succeed)
		}
	}()

	t.Logf("\tWhen get the value of the key %s", key)
	config.GetIntOrPanic(key)
}

func thenGetIntOrPanicMustReturnTheExpectedValue(cases map[string]interface{}, t *testing.T) {
	for k, expected := range cases {
		t.Logf("\tWhen Get the key %s value.", k)
		actualVal := config.GetIntOrPanic(k)
		if actualVal != expected {
			t.Fatalf("\t%s\tFAIL: GetIntOrPanic(%s) expected %d. Actual %d ", infra.Failed, k, expected, actualVal)
		}
		t.Logf("\t%s\tThen the value must be %d", infra.Succeed, expected)
	}
}

func saveConfigurationsInTheEnvironment(configuration map[string]interface{}, t *testing.T) {
	for key, value := range configuration {
		if err := os.Setenv(key, fmt.Sprintf("%v", value)); err != nil {
			t.Fatalf("FAIL: error when saving the key/value %v/%v in the environment: %v", key, value, err)
		}
	}
}

func saveConfigurationsInFile(values map[string]interface{}, t *testing.T) (dir string) {
	tempDir, err := ioutil.TempDir("", "nominee_conf_")
	if err != nil {
		t.Fatalf("FAIL: error when creating temporary dir %v", err)
	}

	var content string
	for k, v := range values {
		content = fmt.Sprintf("%s\n%s=%v", content, k, v)
	}

	envFile := filepath.Join(tempDir, "setting.env")
	err = ioutil.WriteFile(envFile, []byte(content), 0600)
	err = os.Setenv("NOMINEE_CONF_FILE", envFile)
	if err != nil {
		t.Fatalf("FAIL: error when saving the configuration file %v", err)
	}

	return tempDir
}

func tearsDown(configurations map[string]interface{}) {
	config.Reset()
	for key, _ := range configurations {
		_ = os.Unsetenv(key)
	}
}
