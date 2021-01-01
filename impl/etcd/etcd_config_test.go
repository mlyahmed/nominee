package etcd_test

import (
	"context"
	"github/mlyahmed.io/nominee/impl/etcd"
	"github/mlyahmed.io/nominee/infra"
	"github/mlyahmed.io/nominee/pkg/config"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestEtcdConfig_loads_configurations(t *testing.T) {

	t.Logf("Given a valid Etcd configuration")
	{
		for i, example := range validExamples {
			t.Run("", func(t *testing.T) {
				defer tearsDown()
				declareConfigurationExample(example)

				t.Logf("\tTest %d: When load configuration and %s.", i, example.description)
				{
					etcdConfig := etcd.NewEtcdConfig(config.NewBasicConfig())
					etcdConfig.LoadConfig(context.TODO())

					if etcdConfig.Cluster != example.cluster {
						t.Fatalf("\t\t%s FAIL: Config.Cluster, expected <%s> but actual is <%s>", infra.Failed, example.cluster, etcdConfig.Cluster)
					}
					t.Logf("\t\t%s Then the Config.Cluster should be loaded.", infra.Succeed)

					if etcdConfig.Domain != example.domain {
						t.Fatalf("\t\t%s FAIL: Config.Domain, expected <%s> but actual is <%s>", infra.Failed, example.domain, etcdConfig.Domain)
					}
					t.Logf("\t\t%s Then the Config.Domain should be loaded.", infra.Succeed)

					if !reflect.DeepEqual(etcdConfig.Endpoints, strings.Split(example.endpoints, ",")) {
						t.Fatalf("\t\t%s FAIL: Config.Endpoints, expected <%s> but actual is <%s>", infra.Failed, example.endpoints, etcdConfig.Endpoints)
					}
					t.Logf("\t\t%s Then the Config.Endpoints should be loaded.", infra.Succeed)

					if etcdConfig.Username != example.username {
						t.Fatalf("\t\t%s FAIL: Config.Username, expected <%s> but actual is <%s>", infra.Failed, example.username, etcdConfig.Username)
					}
					t.Logf("\t\t%s Then the Config.Username should be loaded.", infra.Succeed)

					if etcdConfig.Password != example.password {
						t.Fatalf("\t\t%s FAIL: Config.Password, expected <%s> but actual is <%s>", infra.Failed, example.password, etcdConfig.Password)
					}
					t.Logf("\t\t%s Then the Config.Password should be loaded.", infra.Succeed)
				}
			})
		}
	}
}

func TestEtcdConfig_panics_when_bad_configuration(t *testing.T) {
	t.Logf("Given an invalid Etcd configuration")
	{
		for i, example := range invalidExamples {
			t.Run("", func(t *testing.T) {
				defer tearsDown()
				declareConfigurationExample(example)

				t.Logf("\tTest %d: When load configuration and %s.", i, example.description)
				{
					defer func() {
						if r := recover(); r == nil {
							t.Fatalf("\t\t%s FAIL: Config.LoadConfig(). Expected the program to panic. Actual not.", infra.Failed)
						} else {
							t.Logf("\t\t%s Then the program must panic.", infra.Succeed)
						}
					}()

					etcdConfig := etcd.NewEtcdConfig(config.NewBasicConfig())
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
