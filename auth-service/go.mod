module github.com/FelipeFelipeRenan/goverse/auth-service

go 1.24.2

require (
	github.com/FelipeFelipeRenan/goverse/common v0.0.0-00010101000000-000000000000
	github.com/FelipeFelipeRenan/goverse/proto v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/uuid v1.6.0
	golang.org/x/oauth2 v0.30.0
)

require (
	github.com/kr/text v0.2.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
)

require (
	cloud.google.com/go/compute/metadata v0.6.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_golang v1.23.2
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
)

require (
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	google.golang.org/grpc v1.73.0
	google.golang.org/protobuf v1.36.8 // indirect
)

replace github.com/FelipeFelipeRenan/goverse/proto => ../proto

replace github.com/FelipeFelipeRenan/goverse/common => ../common
