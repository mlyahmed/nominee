package postgres_test

type configurationExamples struct {
	description        string
	cluster            string
	domain             string
	nodeName           string
	nodeAddress        string
	nodePort           string
	postgresPassword   string
	replicatorUsername string
	replicatorPassword string
}

var validExamples = []configurationExamples{
	{
		description:        "minimum configuration",
		cluster:            "cluster-001",
		domain:             "domain-001",
		nodeName:           "postgres-01",
		nodeAddress:        "node01.postgres.priv",
		postgresPassword:   "postgre$",
		replicatorUsername: "replicator",
		replicatorPassword: "$ecret",
	},
	{
		description:        "full configuration #1",
		cluster:            "cluster-009",
		domain:             "domain-012",
		nodeName:           "postgres-99",
		nodeAddress:        "node99.postgres.priv",
		nodePort:           "5001",
		postgresPassword:   "pg$$$$$",
		replicatorUsername: "repl",
		replicatorPassword: "@$ecret",
	},
	{
		description:        "full configuration #2",
		cluster:            "cluster-209",
		domain:             "domain-713",
		nodeName:           "postgres-77",
		nodeAddress:        "node77.postgres.priv",
		nodePort:           "5000",
		postgresPassword:   "()_+++==MIN$%^&)",
		replicatorUsername: "repl",
		replicatorPassword: "SHUT$$$",
	},
}

var invalidExamples = []configurationExamples{
	{
		description:        "cluster name is missing",
		domain:             "domain-111",
		nodeName:           "node-11",
		nodeAddress:        "node11.postgres.priv",
		postgresPassword:   "postgre$",
		replicatorUsername: "replicator",
		replicatorPassword: "$ecret",
	},
	{
		description:        "domain name is missing",
		cluster:            "cluster-991",
		nodeName:           "postgres-71",
		nodeAddress:        "postgres71.pg.db",
		postgresPassword:   "_-=+&^%$",
		replicatorUsername: "copier",
		replicatorPassword: "^%%$%$7687878",
	},
	{
		description:        "node name is missing",
		cluster:            "cluster-001",
		domain:             "domain-001",
		nodeAddress:        "node01.postgres.priv",
		postgresPassword:   "postgre$",
		replicatorUsername: "replicator",
		replicatorPassword: "$ecret",
	},
	{
		description:        "node address is missing",
		cluster:            "cluster-009",
		domain:             "domain-012",
		nodeName:           "postgres-99",
		postgresPassword:   "pg$$$$$",
		replicatorUsername: "repl",
		replicatorPassword: "@$ecret",
	},
	{
		description:        "postgres password is missing",
		cluster:            "cluster-001",
		domain:             "domain-001",
		nodeName:           "postgres-01",
		nodeAddress:        "node01.postgres.priv",
		replicatorUsername: "replicator",
		replicatorPassword: "$ecret",
	},
	{
		description:        "replicator username is missing",
		cluster:            "cluster-209",
		domain:             "domain-713",
		nodeName:           "postgres-77",
		nodeAddress:        "node77.postgres.priv",
		nodePort:           "5000",
		postgresPassword:   "()_+++==MIN$%^&)",
		replicatorPassword: "SHUT$$$",
	},
	{
		description:        "replicator password is missing",
		cluster:            "cluster-009",
		domain:             "domain-012",
		nodeName:           "postgres-99",
		nodeAddress:        "node99.postgres.priv",
		nodePort:           "5001",
		postgresPassword:   "pg$$$$$",
		replicatorUsername: "repl",
	},
}
