package etcd_test

import (
	"context"
	"github/mlyahmed.io/nominee/impl/etcd"
	"github/mlyahmed.io/nominee/pkg/testutils"
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
					loader := etcd.NewConfigLoader()
					loader.Load(context.TODO())
					etcdConfig := loader.GetSpec()
					if etcdConfig.Cluster != example.cluster {
						t.Fatalf("\t\t%s FAIL: ConfigSpec.Cluster, expected <%s> but actual is <%s>", testutils.Failed, example.cluster, etcdConfig.Cluster)
					}
					t.Logf("\t\t%s Then the ConfigSpec.Cluster should be loaded.", testutils.Succeed)

					if etcdConfig.Domain != example.domain {
						t.Fatalf("\t\t%s FAIL: ConfigSpec.Domain, expected <%s> but actual is <%s>", testutils.Failed, example.domain, etcdConfig.Domain)
					}
					t.Logf("\t\t%s Then the ConfigSpec.Domain should be loaded.", testutils.Succeed)

					if !reflect.DeepEqual(etcdConfig.Endpoints, strings.Split(example.endpoints, ",")) {
						t.Fatalf("\t\t%s FAIL: ConfigSpec.Endpoints, expected <%s> but actual is <%s>", testutils.Failed, example.endpoints, etcdConfig.Endpoints)
					}
					t.Logf("\t\t%s Then the ConfigSpec.Endpoints should be loaded.", testutils.Succeed)

					if etcdConfig.Username != example.username {
						t.Fatalf("\t\t%s FAIL: ConfigSpec.Username, expected <%s> but actual is <%s>", testutils.Failed, example.username, etcdConfig.Username)
					}
					t.Logf("\t\t%s Then the ConfigSpec.Username should be loaded.", testutils.Succeed)

					if etcdConfig.Password != example.password {
						t.Fatalf("\t\t%s FAIL: ConfigSpec.Password, expected <%s> but actual is <%s>", testutils.Failed, example.password, etcdConfig.Password)
					}
					t.Logf("\t\t%s Then the ConfigSpec.Password should be loaded.", testutils.Succeed)
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
							t.Fatalf("\t\t%s FAIL: ConfigSpec.Load(). Expected the program to panic. Actual not.", testutils.Failed)
						} else {
							t.Logf("\t\t%s Then the program must panic.", testutils.Succeed)
						}
					}()

					etcdConfig := etcd.NewConfigLoader()
					etcdConfig.Load(context.TODO())
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
