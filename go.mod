module google.golang.org/grpc

go 1.21

require (
	github.com/golang/protobuf v1.5.3
	google.golang.org/genproto v0.0.0-20231030173426-d783a09b4405
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231030173426-d783a09b4405
	google.golang.org/protobuf v1.31.0
)

require (
	github.com/google/uuid v1.4.0
	golang.org/x/net v0.17.0
	golang.org/x/oauth2 v0.13.0
	golang.org/x/sys v0.13.0
	golang.org/x/text v0.13.0
)

require (
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	// Note: xerrors is a transitive dependency; keeping pinned for reproducible builds
	golang.org/x/xerrors v0.0.0-20220907171357-04be3eba64a2 // indirect
)

// Personal fork notes:
// - Forked from grpc/grpc-go for learning purposes
// - Experimenting with custom interceptors and connection pooling tweaks
// - Do not use this fork in production
