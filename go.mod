module github.com/unee-t/invite

go 1.12

require (
	github.com/apex/log v1.1.0
	github.com/aws/aws-lambda-go v1.11.1
	github.com/aws/aws-sdk-go-v2 v0.9.0
	github.com/go-sql-driver/mysql v1.4.1
	github.com/gorilla/mux v1.7.2
	github.com/hashicorp/go-multierror v1.0.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/kr/pty v1.1.5 // indirect
	github.com/pkg/errors v0.8.1 // indirect
	github.com/prometheus/client_golang v1.0.0
	github.com/prometheus/common v0.6.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.4.2 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/tj/assert v0.0.0-20171129193455-018094318fb0 // indirect
	github.com/tj/go v1.8.6
	github.com/unee-t/env v0.0.0-20190513035325-a55bf10999d5
	golang.org/x/crypto v0.0.0-20190621222207-cc06ce4a13d4 // indirect
	golang.org/x/net v0.0.0-20190620200207-3b0461eec859 // indirect
	golang.org/x/sys v0.0.0-20190621203818-d432491b9138 // indirect
	golang.org/x/tools v0.0.0-20190621195816-6e04913cbbac // indirect
	google.golang.org/appengine v1.6.1 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

replace github.com/satori/go.uuid v1.2.0 => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b

replace github.com/aws/aws-sdk-go-v2 => github.com/aws/aws-sdk-go-v2 v0.7.0
