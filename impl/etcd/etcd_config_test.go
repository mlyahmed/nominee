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

func TestEtcdConfig_it_must_load_all_configurations(t *testing.T) {
	for _, example := range validExamples {
		t.Run("", func(t *testing.T) {
			defer tearsDown()
			declareConfigurationExample(example)
			loader := etcd.NewConfigLoader()
			loader.Load(context.TODO())
			etcdConfig := loader.GetSpec()
			if etcdConfig.Cluster != example.cluster {
				t.Fatalf("\t\t%s FAIL: ConfigSpec.Cluster, expected <%s> but actual is <%s>", testutils.Failed, example.cluster, etcdConfig.Cluster)
			}

			if etcdConfig.Domain != example.domain {
				t.Fatalf("\t\t%s FAIL: ConfigSpec.Domain, expected <%s> but actual is <%s>", testutils.Failed, example.domain, etcdConfig.Domain)
			}

			if !reflect.DeepEqual(etcdConfig.Endpoints, strings.Split(example.endpoints, ",")) {
				t.Fatalf("\t\t%s FAIL: ConfigSpec.Endpoints, expected <%s> but actual is <%s>", testutils.Failed, example.endpoints, etcdConfig.Endpoints)
			}

			if etcdConfig.Username != example.username {
				t.Fatalf("\t\t%s FAIL: ConfigSpec.Username, expected <%s> but actual is <%s>", testutils.Failed, example.username, etcdConfig.Username)
			}

			if etcdConfig.Password != example.password {
				t.Fatalf("\t\t%s FAIL: ConfigSpec.Password, expected <%s> but actual is <%s>", testutils.Failed, example.password, etcdConfig.Password)
			}
		})
	}
}

func TestEtcdConfig_it_must_panic_when_bad_configuration(t *testing.T) {
	for _, example := range invalidExamples {
		t.Run("", func(t *testing.T) {
			defer tearsDown()
			declareConfigurationExample(example)
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("\t\t%s FAIL: ConfigSpec.Load(). Expected the program to panic. Actual not.", testutils.Failed)
				}
			}()
			etcdConfig := etcd.NewConfigLoader()
			etcdConfig.Load(context.TODO())
		})
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
