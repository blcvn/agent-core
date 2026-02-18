module github.com/blcvn/backend/services/ba-persistence-service

go 1.24.0

require (
	gorm.io/driver/postgres v1.5.7
	gorm.io/gorm v1.25.10
	google.golang.org/grpc v1.62.1
	google.golang.org/protobuf v1.33.0
	github.com/blcvn/backend/services/pkg v0.0.0
	github.com/blcvn/backend/services/proto v0.0.0
)

replace (
	github.com/blcvn/backend/services/pkg => ../../services/pkg
	github.com/blcvn/backend/services/proto => ../../services/proto
)
