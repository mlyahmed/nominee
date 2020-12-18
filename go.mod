module github/mlyahmed.io/nominee

go 1.15

require (
	github.com/coreos/bbolt v0.0.0-00010101000000-000000000000 // indirect
	github.com/coreos/etcd v3.3.25+incompatible
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/go-openapi/validate v0.20.0 // indirect
	github.com/go-pg/pg/v10 v10.7.3
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/google/renameio v1.0.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/haproxytech/client-native/v2 v2.1.0
	github.com/haproxytech/config-parser/v2 v2.1.0 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.8.0 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20200427203606-3cfed13b9966 // indirect
	go.etcd.io/etcd v3.3.25+incompatible
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/net v0.0.0-20201216054612-986b41b23924 // indirect
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.5
	github.com/coreos/etcd => github.com/ozonru/etcd v3.3.20-grpc1.27-origmodule+incompatible
	//See https://github.com/etcd-io/etcd/issues/11563#issuecomment-658960252
	google.golang.org/grpc => google.golang.org/grpc v1.27.0
)
