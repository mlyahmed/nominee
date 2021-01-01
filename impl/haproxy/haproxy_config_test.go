package haproxy_test

import (
	"context"
	"github/mlyahmed.io/nominee/impl/haproxy"
	"github/mlyahmed.io/nominee/infra"
	"github/mlyahmed.io/nominee/pkg/config"
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
					haproxyConfig := haproxy.NewHAProxyConfig(config.NewBasicConfig())
					haproxyConfig.LoadConfig(context.TODO())

					if haproxyConfig.Cluster != example.cluster {
						t.Fatalf("\t\t%s FAIL: HAProxyConfig.Cluster, expected <%s> but actual is <%s>", infra.Failed, example.cluster, haproxyConfig.Cluster)
					}
					t.Logf("\t\t%s Then the HAProxyConfig.Cluster should be loaded.", infra.Succeed)

					if haproxyConfig.Domain != example.domain {
						t.Fatalf("\t\t%s FAIL: HAProxyConfig.Domain, expected <%s> but actual is <%s>", infra.Failed, example.domain, haproxyConfig.Domain)
					}
					t.Logf("\t\t%s Then the HAProxyConfig.Domain should be loaded.", infra.Succeed)

					if example.configFile == "" {
						const defaultConfFile = "/usr/local/etc/haproxy/haproxy.cfg"
						if haproxyConfig.ConfigFile != defaultConfFile {
							t.Fatalf("\t\t%s FAIL: HAProxyConfig.ConfigFile, expected the default <%s> but actual is <%s>", infra.Failed, defaultConfFile, haproxyConfig.ConfigFile)
						}
					} else {
						if haproxyConfig.ConfigFile != example.configFile {
							t.Fatalf("\t\t%s FAIL: HAProxyConfig.ConfigFile, expected <%s> but actual is <%s>", infra.Failed, example.configFile, haproxyConfig.ConfigFile)
						}
					}
					t.Logf("\t\t%s Then the HAProxyConfig.ConfigFile should be loaded.", infra.Succeed)

					if example.execFile == "" {
						const defaultExecFile = "/usr/local/sbin/haproxy"
						if haproxyConfig.ExecFile != defaultExecFile {
							t.Fatalf("\t\t%s FAIL: HAProxyConfig.ExecFile, expected the default <%s> but actual is <%s>", infra.Failed, defaultExecFile, haproxyConfig.ExecFile)
						}
					} else {
						if haproxyConfig.ExecFile != example.execFile {
							t.Fatalf("\t\t%s FAIL: HAProxyConfig.ExecFile, expected <%s> but actual is <%s>", infra.Failed, example.execFile, haproxyConfig.ExecFile)
						}
					}
					t.Logf("\t\t%s Then the HAProxyConfig.ExecFile should be loaded.", infra.Succeed)

					if example.txDir == "" {
						const defaultTxDir = "/tmp/haproxy"
						if haproxyConfig.TxDir != defaultTxDir {
							t.Fatalf("\t\t%s FAIL: HAProxyConfig.TxDir, expected the default <%s> but actual is <%s>", infra.Failed, defaultTxDir, haproxyConfig.TxDir)
						}
					} else {
						if haproxyConfig.ExecFile != example.execFile {
							t.Fatalf("\t\t%s FAIL: HAProxyConfig.TxDir, expected <%s> but actual is <%s>", infra.Failed, example.txDir, haproxyConfig.TxDir)
						}
					}
					t.Logf("\t\t%s Then the HAProxyConfig.TxDir should be loaded.", infra.Succeed)
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
							t.Fatalf("\t\t%s FAIL: HAProxyConfig.LoadConfig(). Expected the program to panic. Actual not.", infra.Failed)
						} else {
							t.Logf("\t\t%s Then the program must panic.", infra.Succeed)
						}
					}()

					haproxyConfig := haproxy.NewHAProxyConfig(config.NewBasicConfig())
					haproxyConfig.LoadConfig(context.TODO())
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
