module github.com/taco-labs/taco

go 1.19

require (
	firebase.google.com/go v3.13.0+incompatible
	github.com/aws/aws-sdk-go v1.44.68
	github.com/aws/aws-sdk-go-v2/config v1.15.15
	github.com/aws/aws-sdk-go-v2/service/s3 v1.27.2
	github.com/aws/aws-sdk-go-v2/service/sqs v1.19.10
	github.com/coolsms/coolsms-go v0.0.0-20211005081430-ed879c9fbfd3
	github.com/go-resty/resty/v2 v2.7.0
	github.com/google/uuid v1.3.0
	github.com/labstack/echo/v4 v4.8.0
	github.com/panjf2000/ants/v2 v2.5.0
	github.com/sethvargo/go-envconfig v0.8.2
	github.com/twpayne/go-geom v1.4.3
	github.com/uptrace/bun v1.1.7
	github.com/uptrace/bun/dialect/pgdialect v1.1.7
	github.com/uptrace/bun/driver/pgdriver v1.1.7
	github.com/uptrace/bun/extra/bundebug v1.1.7
	go.uber.org/zap v1.23.0
	gocloud.dev v0.27.0
	golang.org/x/net v0.0.0-20220909164309-bea034e7d591
	golang.org/x/sync v0.0.0-20220907140024-f12130a52804
)

require (
	cloud.google.com/go v0.103.0 // indirect
	cloud.google.com/go/compute v1.7.0 // indirect
	cloud.google.com/go/firestore v1.6.1 // indirect
	cloud.google.com/go/iam v0.3.0 // indirect
	cloud.google.com/go/storage v1.26.0 // indirect
	github.com/aws/aws-sdk-go-v2 v1.16.16 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.4.3 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.12.10 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.9 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.16 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.0.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.9.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.1.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.13.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/sns v1.17.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.11.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.16.10 // indirect
	github.com/aws/smithy-go v1.13.3 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.8 // indirect
	github.com/google/wire v0.5.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.1.0 // indirect
	github.com/googleapis/gax-go/v2 v2.4.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/labstack/gommon v0.3.1 // indirect
	github.com/lib/pq v1.10.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/tmthrgd/go-hex v0.0.0-20190904060850-447a3041c3bc // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.1 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/crypto v0.0.0-20220722155217-630584e8d5aa // indirect
	golang.org/x/oauth2 v0.0.0-20220909003341-f21342109be1 // indirect
	golang.org/x/sys v0.0.0-20220731174439-a90be440212d // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20220722155302-e5dcc9cfc0b9 // indirect
	golang.org/x/xerrors v0.0.0-20220609144429-65e65417b02f // indirect
	google.golang.org/api v0.96.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220810155839-1856144b1d9c // indirect
	google.golang.org/grpc v1.48.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	mellium.im/sasl v0.2.1 // indirect
)

replace gocloud.dev v0.27.0 => ./go-cloud
