package race_test

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/config"
	"github/mlyahmed.io/nominee/pkg/race"
	"os"
	"testing"
)

const (
	succeed = "\u2713"
	failed  = "\u2717"
)

func TestEtcdConfig_loads_configurations(t *testing.T) {

	t.Logf("Given valid ETCD configurations")
	{
		for i, example := range validConfigurationsExamples {
			t.Run("", func(t *testing.T) {
				defer tearsDown()
				declareConfigurationExample(example)

				t.Logf("\tTest %d: When load configuration and %s.", i, example.description)
				{
					etcdConfig := race.NewEtcdConfig(config.NewBasicConfig())
					etcdConfig.LoadConfig(context.TODO())

					if etcdConfig.Cluster != example.cluster {
						t.Fatalf("\t\t%s FAIL: EtcdConfig.Cluster, expected <%s> but actual is <%s>", failed, example.cluster, etcdConfig.Cluster)
					}
					t.Logf("\t\t%s Then the EtcdConfig.Cluster should be loaded.", succeed)

					if etcdConfig.Domain != example.domain {
						t.Fatalf("\t\t%s FAIL: EtcdConfig.Domain, expected <%s> but actual is <%s>", failed, example.domain, etcdConfig.Domain)
					}
					t.Logf("\t\t%s Then the EtcdConfig.Domain should be loaded.", succeed)

					if etcdConfig.Endpoints != example.endpoints {
						t.Fatalf("\t\t%s FAIL: EtcdConfig.Endpoints, expected <%s> but actual is <%s>", failed, example.endpoints, etcdConfig.Endpoints)
					}
					t.Logf("\t\t%s Then the EtcdConfig.Endpoints should be loaded.", succeed)

					if etcdConfig.Username != example.username {
						t.Fatalf("\t\t%s FAIL: EtcdConfig.Username, expected <%s> but actual is <%s>", failed, example.username, etcdConfig.Username)
					}
					t.Logf("\t\t%s Then the EtcdConfig.Username should be loaded.", succeed)

					if etcdConfig.Password != example.password {
						t.Fatalf("\t\t%s FAIL: EtcdConfig.Password, expected <%s> but actual is <%s>", failed, example.password, etcdConfig.Password)
					}
					t.Logf("\t\t%s Then the EtcdConfig.Password should be loaded.", succeed)
				}
			})
		}
	}
}

func TestEtcdConfig_panics_when_bad_configuration(t *testing.T) {
	t.Logf("Given invalid ETCD configurations")
	{
		for i, example := range invalidConfigurationExamples {
			t.Run("", func(t *testing.T) {
				defer tearsDown()
				declareConfigurationExample(example)

				t.Logf("\tTest %d: When load configuration and %s.", i, example.description)
				{
					defer func() {
						if r := recover(); r == nil {
							t.Fatalf("\t\t%s FAIL: EtcdConfig.LoadConfig(). Expected the program to panic. Actual not.", failed)
						} else {
							t.Logf("\t\t%s Then the program must panic.", succeed)
						}
					}()

					etcdConfig := race.NewEtcdConfig(config.NewBasicConfig())
					etcdConfig.LoadConfig(context.TODO())
				}

			})
		}
	}
}

func declareConfigurationExample(example configurationExample) {
	_ = os.Setenv("NOMINEE_CLUSTER_NAME", example.cluster)
	_ = os.Setenv("NOMINEE_DOMAIN_NAME", example.domain)
	_ = os.Setenv("NOMINEE_ETCD_ENDPOINTS", example.endpoints)
	_ = os.Setenv("NOMINEE_ETCD_USERNAME", example.username)
	_ = os.Setenv("NOMINEE_ETCD_PASSWORD", example.password)
}

func tearsDown() {
	_ = os.Unsetenv("NOMINEE_CLUSTER_NAME")
	_ = os.Unsetenv("NOMINEE_DOMAIN_NAME")
	_ = os.Unsetenv("NOMINEE_ETCD_ENDPOINTS")
	_ = os.Unsetenv("NOMINEE_ETCD_USERNAME")
	_ = os.Unsetenv("NOMINEE_ETCD_PASSWORD")
}
