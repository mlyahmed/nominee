package proxy_test

type configurationExample struct {
	description string
	cluster     string
	domain      string
	configFile  string
	execFile    string
	txDir       string
}

var validExamples = []configurationExample{
	{
		description: "minimum configuration",
		cluster:     "cluster-001",
		domain:      "domain-002",
	},
	{
		description: "full configuration #1",
		cluster:     "cluster-111",
		domain:      "domain-542",
		configFile:  "/etc/haproxy/haproxy.cfg",
		execFile:    "/bin/haproxy",
		txDir:       "/usr/local/tmp/haproxy",
	},
	{
		description: "full configuration #2",
		cluster:     "cluster-65888",
		domain:      "domain/6565659",
		configFile:  "/usr/local/etc/haproxy/haproxy.cfg",
		execFile:    "/usr/local/sbin/haproxy",
		txDir:       "/usr/tmp/haproxy",
	},
}

var invalidExamples = []configurationExample{
	{
		description: "cluster name is missing",
		domain:      "domain-002",
	},
	{
		description: "domain name is missing",
		cluster:     "cluster",
	},
}
