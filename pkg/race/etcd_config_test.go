package race_test

import (
	"context"
	"github/mlyahmed.io/nominee/pkg/config"
	"github/mlyahmed.io/nominee/pkg/race"
	"github/mlyahmed.io/nominee/pkg/testutils"
	"os"
	"testing"
)

func TestEtcdConfig_loads_configurations(t *testing.T) {

	t.Logf("Given valid Etcd configurations")
	{
		for i, example := range validExamples {
			t.Run("", func(t *testing.T) {
				defer tearsDown()
				declareConfigurationExample(example)

				t.Logf("\tTest %d: When load configuration and %s.", i, example.description)
				{
					etcdConfig := race.NewEtcdConfig(config.NewBasicConfig())
					etcdConfig.LoadConfig(context.TODO())

					if etcdConfig.Cluster != example.cluster {
						t.Fatalf("\t\t%s FAIL: EtcdConfig.Cluster, expected <%s> but actual is <%s>", testutils.Failed, example.cluster, etcdConfig.Cluster)
					}
					t.Logf("\t\t%s Then the EtcdConfig.Cluster should be loaded.", testutils.Succeed)

					if etcdConfig.Domain != example.domain {
						t.Fatalf("\t\t%s FAIL: EtcdConfig.Domain, expected <%s> but actual is <%s>", testutils.Failed, example.domain, etcdConfig.Domain)
					}
					t.Logf("\t\t%s Then the EtcdConfig.Domain should be loaded.", testutils.Succeed)

					if etcdConfig.Endpoints != example.endpoints {
						t.Fatalf("\t\t%s FAIL: EtcdConfig.Endpoints, expected <%s> but actual is <%s>", testutils.Failed, example.endpoints, etcdConfig.Endpoints)
					}
					t.Logf("\t\t%s Then the EtcdConfig.Endpoints should be loaded.", testutils.Succeed)

					if etcdConfig.Username != example.username {
						t.Fatalf("\t\t%s FAIL: EtcdConfig.Username, expected <%s> but actual is <%s>", testutils.Failed, example.username, etcdConfig.Username)
					}
					t.Logf("\t\t%s Then the EtcdConfig.Username should be loaded.", testutils.Succeed)

					if etcdConfig.Password != example.password {
						t.Fatalf("\t\t%s FAIL: EtcdConfig.Password, expected <%s> but actual is <%s>", testutils.Failed, example.password, etcdConfig.Password)
					}
					t.Logf("\t\t%s Then the EtcdConfig.Password should be loaded.", testutils.Succeed)
				}
			})
		}
	}
}

func TestEtcdConfig_panics_when_bad_configuration(t *testing.T) {
	t.Logf("Given invalid Etcd configurations")
	{
		for i, example := range invalidExamples {
			t.Run("", func(t *testing.T) {
				defer tearsDown()
				declareConfigurationExample(example)

				t.Logf("\tTest %d: When load configuration and %s.", i, example.description)
				{
					defer func() {
						if r := recover(); r == nil {
							t.Fatalf("\t\t%s FAIL: EtcdConfig.LoadConfig(). Expected the program to panic. Actual not.", testutils.Failed)
						} else {
							t.Logf("\t\t%s Then the program must panic.", testutils.Succeed)
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
