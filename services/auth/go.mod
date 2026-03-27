module auth

go 1.25.5

require (
	github.com/Dimassin/articles-microservices/proto/auth v0.0.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/lib/pq v1.12.0
	golang.org/x/crypto v0.49.0
	google.golang.org/grpc v1.79.3
)

require (
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/segmentio/kafka-go v0.4.50 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
	golang.org/x/text v0.35.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

replace github.com/Dimassin/articles-microservices/proto/auth => ../../proto/auth
