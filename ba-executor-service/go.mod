module github.com/blcvn/backend/services/ba-executor-service

go 1.24.0

require (
	google.golang.org/grpc v1.62.1
	google.golang.org/protobuf v1.33.0
	github.com/blcvn/kratos-proto/go/ba-agent v1.1.0
	github.com/blcvn/backend/services/pkg v0.0.0
	github.com/blcvn/backend/services/ba-mcp-server v0.0.0
	github.com/blcvn/backend/services/proto v0.0.0
)

replace (
	github.com/blcvn/backend/services/pkg => ../../services/pkg
	github.com/blcvn/backend/services/ba-mcp-server => ../../services/ba-mcp-server
	github.com/blcvn/backend/services/proto => ../../services/proto
)
