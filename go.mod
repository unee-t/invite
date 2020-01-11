module github.com/unee-t-ins/invite

go 1.12

require (
	github.com/apex/log v1.1.1
	github.com/aws/aws-lambda-go v1.13.3
	github.com/aws/aws-sdk-go-v2 v0.18.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/go-multierror v1.0.0
	github.com/kr/pretty v0.1.0 // indirect
	github.com/prometheus/client_golang v1.3.0
	github.com/satori/go.uuid v1.2.0
	github.com/tj/go v1.8.6
	github.com/unee-t-ins/env v0.2.2
	golang.org/x/sys v0.0.0-20200107162124-548cf772de50 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/satori/go.uuid v1.2.0 => github.com/satori/go.uuid v1.2.1-0.20181028125025-b2ce2384e17b
