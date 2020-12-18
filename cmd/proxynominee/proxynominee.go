package main

import (
	"fmt"
	"github.com/haproxytech/client-native/v2/configuration"
	_ "github.com/haproxytech/client-native/v2/runtime"
)

func main() {
	client := &configuration.Client{}
	confParams := configuration.ClientParams{
		ConfigurationFile:      "/home/ahmed/data/projects/postgres-operator/labs/nominee/haproxy.cfg",
		Haproxy:                "/usr/sbin/haproxy",
		UseValidation:          true,
		PersistentTransactions: true,
		TransactionDir:         "/tmp/haproxy",
	}

	if err := client.Init(confParams); err != nil {
		panic(err)
	}

	version, backends, err := client.GetBackends("")
	if err != nil {
		panic(err)
	}
	fmt.Println(version)
	for _, v := range backends {
		fmt.Sprintln(v.Name)
	}
}
