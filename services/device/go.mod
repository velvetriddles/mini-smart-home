module github.com/velvetriddles/mini-smart-home/services/device

go 1.23.0

toolchain go1.23.2

require (
	github.com/eclipse/paho.mqtt.golang v1.4.3
	github.com/google/uuid v1.6.0
	github.com/prometheus/client_golang v1.18.0
	github.com/segmentio/kafka-go v0.4.47
	github.com/stretchr/testify v1.9.0
	github.com/velvetriddles/mini-smart-home/proto v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.71.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/matttproud/golang_protobuf_extensions/v2 v2.0.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.18 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.45.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250414145226-207652e42e2e // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/velvetriddles/mini-smart-home/proto => ../../proto
