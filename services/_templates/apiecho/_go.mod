module <<<Module>>>

go 1.21.1

require (
	github.com/brpaz/echozap v1.1.3
	github.com/getkin/kin-openapi v0.122.0
	github.com/labstack/echo/v4 v4.11.4
	github.com/oapi-codegen/echo-middleware v1.0.1
	github.com/oapi-codegen/runtime v1.1.1
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.8.4
	gitlab.com/ginfra/ginfra v0.1.0
	go.uber.org/zap v1.25.0

)

// For development time.  This is for when you are working on ginfra at the same time.  You will need to build
// ginfra at least once on the machine before this works.
replace gitlab.com/ginfra/ginfra v0.1.0 => <<<GINFRA_HOME>>>

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/cbroglie/mustache v1.4.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-openapi/jsonpointer v0.20.0 // indirect
	github.com/go-openapi/swag v0.22.4 // indirect
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/google/uuid v1.5.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/invopop/yaml v0.2.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect; indirec
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/perimeterx/marshmallow v1.1.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.12.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
