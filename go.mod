module email-sender

go 1.22

toolchain go1.24.5

require (
	github.com/aws/aws-sdk-go v1.44.224
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-redis/redis/v8 v8.11.5
	github.com/hibiken/asynq v0.25.0
	github.com/joho/godotenv v1.5.1 // Add this for .env support
	github.com/valyala/fasthttp v1.47.0
	go.mongodb.org/mongo-driver v1.10.0
	golang.org/x/crypto v0.7.0
	golang.org/x/net v0.8.0
)

require (
	github.com/aws/aws-lambda-go v1.49.0
	github.com/google/uuid v1.6.0
	github.com/valyala/fasthttprouter v0.0.0-20160217050331-24073dd8f323
)

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.16.3 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/redis/go-redis/v9 v9.7.0 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.1 // indirect
	github.com/xdg-go/stringprep v1.0.3 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	golang.org/x/time v0.7.0 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
)
