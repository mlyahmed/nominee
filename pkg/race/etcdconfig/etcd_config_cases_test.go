package etcdconfig_test

type configurationExample struct {
	description string
	cluster     string
	domain      string
	endpoints   string
	username    string
	password    string
}

var validExamples = []configurationExample{
	{
		description: "Username/password are empty.",
		cluster:     "nominee",
		domain:      "postgres",
		endpoints:   "etcd-1:2378,etcd-2:2378,etcd-3:2378",
	},
	{
		description: "full configuration",
		cluster:     "cluster-002",
		domain:      "foo",
		endpoints:   "192.168.0.1:2378,192.168.0.2:2378,192.168.0.13:2378",
		username:    "configure",
		password:    "confi9ure",
	},
	{
		description: "full configuration",
		cluster:     "cluster-003",
		domain:      "domain-007",
		endpoints:   "node1.config.priv,node2.config.priv,node3.config.priv",
		username:    "config",
		password:    "configXXX",
	},
}

var invalidExamples = []configurationExample{
	{
		description: "the cluster name is missing",
		domain:      "postgres",
		endpoints:   "etcd-1:2378,etcd-2:2378,etcd-3:2378",
	},
	{
		description: "the domain name is missing",
		cluster:     "nominee",
		endpoints:   "etcd-1:2378,etcd-2:2378,etcd-3:2378",
	},
	{
		description: "the endpoints is missing",
		cluster:     "nominee",
		domain:      "postgres",
	},
}
