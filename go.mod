module github.com/cloud-barista/cb-ladybug

go 1.15

replace (
	github.com/coreos/bbolt => go.etcd.io/bbolt v1.3.3
	github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.0.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/beego/beego/v2 v2.0.1
	github.com/bramvdbogaerde/go-scp v0.0.0-20200518191442-5c8efdd1d925 // indirect
	github.com/cloud-barista/cb-spider v0.3.0-espresso
	github.com/cloud-barista/cb-store v0.3.0-espresso
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/go-openapi/spec v0.19.9 // indirect
	github.com/go-openapi/swag v0.19.9 // indirect
	github.com/go-resty/resty/v2 v2.3.0
	github.com/google/uuid v1.1.2
	github.com/labstack/echo/v4 v4.1.17
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/shiena/ansicolor v0.0.0-20200904210342-c7312218db18 // indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/swaggo/echo-swagger v1.0.0
	github.com/swaggo/swag v1.6.7
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0 // indirect
	golang.org/x/sys v0.0.0-20201005065044-765f4ea38db3 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
)
