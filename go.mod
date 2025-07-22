module gitlab.com/warrant1/warrant/chain-xrpl

go 1.24

toolchain go1.24.5

require (
	github.com/google/wire v0.6.0
	github.com/spf13/cobra v1.9.1
	gitlab.com/warrant1/warrant/protobuf v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.73.0
)

replace gitlab.com/warrant1/warrant/protobuf => ./proto

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250324211829-b45e905df463 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
