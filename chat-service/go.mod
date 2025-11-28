module github.com/FelipeFelipeRenan/goverse/chat-service

go 1.24.2

toolchain go1.24.6

require (
	github.com/FelipeFelipeRenan/goverse/common v0.0.0-00010101000000-000000000000
	github.com/FelipeFelipeRenan/goverse/proto v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/gorilla/websocket v1.5.3
	github.com/jackc/pgx/v5 v5.7.6
	google.golang.org/grpc v1.76.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/segmentio/kafka-go v0.4.49 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/protobuf v1.36.8 // indirect
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/redis/go-redis/v9 v9.16.0
	golang.org/x/crypto v0.40.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)

replace github.com/FelipeFelipeRenan/goverse/proto => ../proto

replace github.com/FelipeFelipeRenan/goverse/common => ../common
