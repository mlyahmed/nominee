package config_test

import (
	"fmt"
	"github/mlyahmed.io/nominee/pkg/config"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

const succeed = "\u2713"
const failed = "\u2717"

func TestGetStringOrPanic_panics_when_get_empty_string_value(t *testing.T) {
	missingKeys := map[string]string{"K1": "", "K2": "", "K3": ""}
	t.Run("Environment", func(t *testing.T) {
		t.Logf("Given the keys %s are loaded with empty values in the environment.", missingKeys)
		{
			for key, _ := range missingKeys {
				_ = os.Setenv(key, "")
			}

			for _, key := range missingKeys {
				thenGetStringOrPanicMustPanic(key, t)
			}
		}
	})

	t.Run("File", func(t *testing.T) {
		t.Logf("Given the keys %s are loaded with empty values in the file.", missingKeys)
		{
			dir := saveInConfigurationFile(missingKeys, t)
			defer os.RemoveAll(dir)
			for _, key := range missingKeys {
				thenGetStringOrPanicMustPanic(key, t)
			}
		}
	})
}

func TestGetStringOrPanic_panics_when_get_missing_string_value(t *testing.T) {
	missingKeys := []string{"K1", "K2", "K3"}
	t.Logf("Given these missing keys %s.", missingKeys)
	{
		for _, key := range missingKeys {
			thenGetStringOrPanicMustPanic(key, t)
		}
	}
}

func TestGetStringOrPanic_returns_value_when_the_key_is_defined(t *testing.T) {
	cases := map[string]string{"K1": "V1", "K2": "V2", "K3": "V3"}

	t.Run("Environment", func(t *testing.T) {
		t.Logf("Given the searshed keys/values %s are loaded in the environment", cases)
		{
			for k, v := range cases {
				config.SetDefault(k, "THE DEFAULT VALUE THAT MUST BE IGNORED")
				_ = os.Setenv(k, v)
			}

			thenGetStringOrPanicMustReturnTheExpectedValue(t, cases)
		}
	})

	t.Run("File", func(t *testing.T) {
		t.Logf("Given the searshed keys/values %s are loaded in the file.", cases)
		{
			dir := saveInConfigurationFile(cases, t)
			defer os.RemoveAll(dir)
			for k, _ := range cases {
				config.SetDefault(k, "THE DEFAULT VALUE THAT MUST BE IGNORED")
			}

			thenGetStringOrPanicMustReturnTheExpectedValue(t, cases)
		}
	})

}

func TestGetStringOrPanic_returns_the_default_Value(t *testing.T) {
	defaultValues := map[string]string{"K1": "D1", "K2": "D2", "K3": "D3"}
	t.Logf("Given the searshed keys/defaults %s are declared.", defaultValues)
	{
		for k, v := range defaultValues {
			_ = os.Unsetenv(k)
			config.SetDefault(k, v)
		}

		thenGetStringOrPanicMustReturnTheExpectedValue(t, defaultValues)
	}
}

func thenGetStringOrPanicMustPanic(key string, t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("\t%s\tFAIL: GetStringOrPanic(%s). Expected the program to panic. Actual not.", failed, key)
		} else {
			t.Logf("\t%s\tThen the program must panic.", succeed)
		}
	}()

	t.Logf("\tWhen get the value of the key %s", key)
	config.GetStringOrPanic(key)
}

func thenGetStringOrPanicMustReturnTheExpectedValue(t *testing.T, cases map[string]string) {
	for k, expected := range cases {
		t.Logf("\tWhen Get the key %s value.", k)
		actualVal := config.GetStringOrPanic(k)
		if actualVal != expected {
			t.Fatalf("\t%s\tFAIL: GetStringOrPanic(%s) expected %s. Actual %s ", failed, k, expected, actualVal)
		}
		t.Logf("\t%s\tThen the value must be %s", succeed, expected)
	}
}

func TestGetString_returns_empty_string_when_the_key_is_not_defined(t *testing.T) {
	emptyValues := map[string]string{"A": "", "P": "", "C": ""}
	t.Logf("Given the missing keys %s.", emptyValues)
	{
		thenGetStringMustReturnTheExpectedValue(emptyValues, t)
	}
}

func TestGetString_returns_the_default_value(t *testing.T) {
	defaultValues := map[string]string{"K1": "A", "K2": "B", "K3": "C"}
	t.Logf("Given the searshed keys/defaults %s are declared.", defaultValues)
	{
		for k, v := range defaultValues {
			_ = os.Unsetenv(k)
			config.SetDefault(k, v)
		}

		thenGetStringMustReturnTheExpectedValue(defaultValues, t)
	}
}

func TestGetString_returns_the_defined_value(t *testing.T) {
	cases := map[string]string{"X1": "V1", "X2": "V2", "X3": "V3"}

	t.Run("Environment", func(t *testing.T) {
		t.Logf("Given the searshed keys/values %s are loaded in the environment.", cases)
		{
			for k, v := range cases {
				config.SetDefault(k, "THE DEFAULT VALUE THAT MUST BE IGNORED")
				_ = os.Setenv(k, v)
			}

			thenGetStringMustReturnTheExpectedValue(cases, t)
		}
	})

	t.Run("File", func(t *testing.T) {
		t.Logf("Given the searshed keys/values %s are loaded in a configuration file.", cases)
		{
			dir := saveInConfigurationFile(cases, t)
			defer os.RemoveAll(dir)

			for k, _ := range cases {
				config.SetDefault(k, "THE DEFAULT VALUE THAT MUST BE IGNORED")
			}

			thenGetStringMustReturnTheExpectedValue(cases, t)
		}
	})
}

func thenGetStringMustReturnTheExpectedValue(cases map[string]string, t *testing.T) {
	for k, expected := range cases {
		t.Logf("\tWhen Get the key %s value.", k)
		actualVal := config.GetString(k)
		if actualVal != expected {
			t.Fatalf("\t%s\tFAIL: GetString(%s) expected %s. Actual %s ", failed, k, expected, actualVal)
		}
		t.Logf("\t%s\tThen the value must be %s", succeed, expected)
	}
}

func saveInConfigurationFile(values map[string]string, t *testing.T) (dir string) {
	tempDir, err := ioutil.TempDir("", "nominee_conf_")
	if err != nil {
		t.Fatalf("FAIL: error when creating temporary dir %v", err)
	}

	var content string
	for k, v := range values {
		content = fmt.Sprintf("%s\n%s=%s", content, k, v)
	}

	envFile := filepath.Join(tempDir, "setting.env")
	err = ioutil.WriteFile(envFile, []byte(content), 0600)
	err = os.Setenv("NOMINEE_CONF_FILE", envFile)
	if err != nil {
		t.Fatalf("FAIL: error when defining the configuration file %v", err)
	}

	return tempDir
}
