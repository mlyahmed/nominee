package postgres_test

import (
	"context"
	"github/mlyahmed.io/nominee/impl/postgres"
	"github/mlyahmed.io/nominee/infra"
	"os"
	"strconv"
	"testing"
)

func TestPGConfig_loads_configurations(t *testing.T) {
	t.Logf("Given valid Postgres configurations")
	{
		for i, example := range validExamples {
			t.Run("", func(t *testing.T) {
				defer tearsDown()
				declareConfigurationExample(example)

				t.Logf("\tTest %d: When load configuration and %s.", i, example.description)
				{
					loader := postgres.NewConfigLoader()
					loader.Load(context.TODO())
					pgConfig := loader.GetSpec()
					if pgConfig.Cluster != example.cluster {
						t.Fatalf("\t\t%s FAIL: ConfigSpec.Cluster, expected <%s> but actual is <%s>", infra.Failed, example.cluster, pgConfig.Cluster)
					}
					t.Logf("\t\t%s Then the ConfigSpec.Cluster should be loaded.", infra.Succeed)

					if pgConfig.Domain != example.domain {
						t.Fatalf("\t\t%s FAIL: ConfigSpec.Domain, expected <%s> but actual is <%s>", infra.Failed, example.domain, pgConfig.Domain)
					}
					t.Logf("\t\t%s Then the ConfigSpec.Domain should be loaded.", infra.Succeed)

					if pgConfig.NodeSpec.Name != example.nodeName {
						t.Fatalf("\t\t%s FAIL: ConfigSpec.GetSpec.GetName, expected <%s> but actual is <%s>", infra.Failed, example.nodeName, pgConfig.NodeSpec.Name)
					}
					t.Logf("\t\t%s Then the ConfigSpec.GetSpec.GetName should be loaded.", infra.Succeed)

					if pgConfig.NodeSpec.Address != example.nodeAddress {
						t.Fatalf("\t\t%s FAIL: ConfigSpec.GetSpec.GetAddress, expected <%s> but actual is <%s>", infra.Failed, example.nodeAddress, pgConfig.NodeSpec.Address)
					}
					t.Logf("\t\t%s Then the ConfigSpec.GetSpec.GetAddress should be loaded.", infra.Succeed)

					if example.nodePort != "" {
						if strconv.Itoa(int(pgConfig.NodeSpec.Port)) != example.nodePort {
							t.Fatalf("\t\t%s FAIL: ConfigSpec.GetSpec.Port, expected <%s> but actual is <%d>", infra.Failed, example.nodePort, pgConfig.NodeSpec.Port)
						}
					} else {
						if pgConfig.NodeSpec.Port != 5432 {
							t.Fatalf("\t\t%s FAIL: ConfigSpec.GetSpec.Port, expected default port number 5432 but actual is <%d>", infra.Failed, pgConfig.NodeSpec.Port)
						}
					}
					t.Logf("\t\t%s Then the pgConfig.GetSpec.Port should be loaded.", infra.Succeed)

					if pgConfig.Postgres.Password != example.postgresPassword {
						t.Fatalf("\t\t%s FAIL: ConfigSpec.Postgres.Password, expected <%s> but actual is <%s>", infra.Failed, example.postgresPassword, pgConfig.Postgres.Password)
					}
					t.Logf("\t\t%s Then the ConfigSpec.Postgres.Password should be loaded.", infra.Succeed)

					envPassword := os.Getenv("POSTGRES_PASSWORD")
					if envPassword != example.postgresPassword {
						t.Fatalf("\t\t%s FAIL: Getenv('POSTGRES_PASSWORD'), expected <%s> but actual is <%s>", infra.Failed, example.postgresPassword, envPassword)
					}
					t.Logf("\t\t%s Then the env. variable POSTGRES_PASSWORD should be loaded.", infra.Succeed)

					if pgConfig.Replicator.Username != example.replicatorUsername {
						t.Fatalf("\t\t%s FAIL: ConfigSpec.Replicator.Username, expected <%s> but actual is <%s>", infra.Failed, example.replicatorUsername, pgConfig.Replicator.Username)
					}
					t.Logf("\t\t%s Then the ConfigSpec.Replicator.Username should be loaded.", infra.Succeed)

					if pgConfig.Replicator.Password != example.replicatorPassword {
						t.Fatalf("\t\t%s FAIL: ConfigSpec.Replicator.Password, expected <%s> but actual is <%s>", infra.Failed, example.replicatorPassword, pgConfig.Replicator.Password)
					}
					t.Logf("\t\t%s Then the ConfigSpec.Replicator.Password should be loaded.", infra.Succeed)
				}

			})
		}
	}
}

func TestEtcdConfig_panics_when_bad_configuration(t *testing.T) {
	t.Logf("Given invalid Postgres configurations")
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

					pgConfig := postgres.NewConfigLoader()
					pgConfig.Load(context.TODO())
				}

			})
		}
	}
}

func declareConfigurationExample(example configurationExamples) {
	_ = os.Setenv("NOMINEE_CLUSTER_NAME", example.cluster)
	_ = os.Setenv("NOMINEE_DOMAIN_NAME", example.domain)
	_ = os.Setenv("NOMINEE_POSTGRES_NODE_NAME", example.nodeName)
	_ = os.Setenv("NOMINEE_POSTGRES_NODE_ADDRESS", example.nodeAddress)
	_ = os.Setenv("NOMINEE_POSTGRES_NODE_PORT", example.nodePort)
	_ = os.Setenv("NOMINEE_POSTGRES_PASSWORD", example.postgresPassword)
	_ = os.Setenv("NOMINEE_POSTGRES_REP_USERNAME", example.replicatorUsername)
	_ = os.Setenv("NOMINEE_POSTGRES_REP_PASSWORD", example.replicatorPassword)
}

func tearsDown() {
	_ = os.Unsetenv("NOMINEE_CLUSTER_NAME")
	_ = os.Unsetenv("NOMINEE_DOMAIN_NAME")
	_ = os.Unsetenv("NOMINEE_POSTGRES_NODE_NAME")
	_ = os.Unsetenv("NOMINEE_POSTGRES_NODE_ADDRESS")
	_ = os.Unsetenv("NOMINEE_POSTGRES_NODE_PORT")
	_ = os.Unsetenv("NOMINEE_POSTGRES_PASSWORD")
	_ = os.Unsetenv("NOMINEE_POSTGRES_REP_USERNAME")
	_ = os.Unsetenv("NOMINEE_POSTGRES_REP_PASSWORD")
}
