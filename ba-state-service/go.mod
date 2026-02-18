module github.com/blcvn/backend/services/ba-state-service

go 1.24.0

require (
	github.com/blcvn/backend/services/pkg v0.0.0
	github.com/blcvn/backend/services/proto v0.0.0
	google.golang.org/grpc v1.62.1
	google.golang.org/protobuf v1.33.0
)

replace (
	github.com/blcvn/backend/services/pkg => ../../services/pkg
	github.com/blcvn/backend/services/proto => ../../services/proto
)
