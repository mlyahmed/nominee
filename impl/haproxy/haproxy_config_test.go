package haproxy_test

import (
	"context"
	"github/mlyahmed.io/nominee/impl/haproxy"
	"github/mlyahmed.io/nominee/infra"
	"os"
	"testing"
)

func TestHAProxyConfig_loads_configurations(t *testing.T) {

	t.Logf("Given a valid HAProxy configuration")
	{
		for i, example := range validExamples {
			t.Run("", func(t *testing.T) {
				defer tearsDown()
				declareConfigurationExample(example)

				t.Logf("\tTest %d: When load configuration and %s.", i, example.description)
				{
					loader := haproxy.NewConfigLoader()
					loader.Load(context.TODO())
					config := loader.GetSpec()

					if config.Cluster != example.cluster {
						t.Fatalf("\t\t%s FAIL: ConfigSpec.Cluster, expected <%s> but actual is <%s>", infra.Failed, example.cluster, config.Cluster)
					}
					t.Logf("\t\t%s Then the ConfigSpec.Cluster should be loaded.", infra.Succeed)

					if config.Domain != example.domain {
						t.Fatalf("\t\t%s FAIL: ConfigSpec.Domain, expected <%s> but actual is <%s>", infra.Failed, example.domain, config.Domain)
					}
					t.Logf("\t\t%s Then the ConfigSpec.Domain should be loaded.", infra.Succeed)

					if example.configFile == "" {
						const defaultConfFile = "/usr/local/etc/haproxy/haproxy.cfg"
						if config.ConfigFile != defaultConfFile {
							t.Fatalf("\t\t%s FAIL: ConfigSpec.ConfigFile, expected the default <%s> but actual is <%s>", infra.Failed, defaultConfFile, config.ConfigFile)
						}
					} else {
						if config.ConfigFile != example.configFile {
							t.Fatalf("\t\t%s FAIL: ConfigSpec.ConfigFile, expected <%s> but actual is <%s>", infra.Failed, example.configFile, config.ConfigFile)
						}
					}
					t.Logf("\t\t%s Then the ConfigSpec.ConfigFile should be loaded.", infra.Succeed)

					if example.execFile == "" {
						const defaultExecFile = "/usr/local/sbin/haproxy"
						if config.ExecFile != defaultExecFile {
							t.Fatalf("\t\t%s FAIL: ConfigSpec.ExecFile, expected the default <%s> but actual is <%s>", infra.Failed, defaultExecFile, config.ExecFile)
						}
					} else {
						if config.ExecFile != example.execFile {
							t.Fatalf("\t\t%s FAIL: ConfigSpec.ExecFile, expected <%s> but actual is <%s>", infra.Failed, example.execFile, config.ExecFile)
						}
					}
					t.Logf("\t\t%s Then the ConfigSpec.ExecFile should be loaded.", infra.Succeed)

					if example.txDir == "" {
						const defaultTxDir = "/tmp/haproxy"
						if config.TxDir != defaultTxDir {
							t.Fatalf("\t\t%s FAIL: ConfigSpec.TxDir, expected the default <%s> but actual is <%s>", infra.Failed, defaultTxDir, config.TxDir)
						}
					} else {
						if config.ExecFile != example.execFile {
							t.Fatalf("\t\t%s FAIL: ConfigSpec.TxDir, expected <%s> but actual is <%s>", infra.Failed, example.txDir, config.TxDir)
						}
					}
					t.Logf("\t\t%s Then the ConfigSpec.TxDir should be loaded.", infra.Succeed)
				}
			})
		}
	}
}

func TestHAProxyConfig_panics_when_bad_configuration(t *testing.T) {
	t.Logf("Given an invalid HAProxy configuration")
	{
		for i, example := range invalidExamples {
			t.Run("", func(t *testing.T) {
				defer tearsDown()
				declareConfigurationExample(example)

				t.Logf("\tTest %d: When load configuration and %s.", i, example.description)
				{
					defer func() {
						if r := recover(); r == nil {
							t.Fatalf("\t\t%s FAIL: ConfigSpec.Load(). Expected the program to panic. Actual not.", infra.Failed)
						} else {
							t.Logf("\t\t%s Then the program must panic.", infra.Succeed)
						}
					}()

					haproxyConfig := haproxy.NewConfigLoader()
					haproxyConfig.Load(context.TODO())
				}

			})
		}
	}
}

func declareConfigurationExample(example configurationExample) {
	_ = os.Setenv("NOMINEE_CLUSTER_NAME", example.cluster)
	_ = os.Setenv("NOMINEE_DOMAIN_NAME", example.domain)
	_ = os.Setenv("NOMINEE_HAPROXY_CONFIG_FILE", example.configFile)
	_ = os.Setenv("NOMINEE_HAPROXY_EXEC_FILE", example.execFile)
	_ = os.Setenv("NOMINEE_HAPROXY_TX_DIR", example.txDir)
}

func tearsDown() {
	_ = os.Unsetenv("NOMINEE_CLUSTER_NAME")
	_ = os.Unsetenv("NOMINEE_DOMAIN_NAME")
	_ = os.Unsetenv("NOMINEE_HAPROXY_CONFIG_FILE")
	_ = os.Unsetenv("NOMINEE_HAPROXY_EXEC_FILE")
	_ = os.Unsetenv("NOMINEE_HAPROXY_TX_DIR")
}
