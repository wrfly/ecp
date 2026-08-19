# ecp

> Environment config parser

If you run your application in a container and deploy it via a docker-compose file, then you may need this tool
for parsing configuration easily instead of mounting an external config file. You can simply set some environments
and then `ecp` will help you fill the configs. Or, you can `COPY` a "default" config file to the image, and change some
variables by overwriting the keys via environments.

The environment config keys can be auto generated or set by the `yaml` or `env` tag.

The only thing you should do is importing this package, and `Parse` your config.

[![Go Report Card](https://goreportcard.com/badge/github.com/wrfly/ecp)](https://goreportcard.com/report/github.com/wrfly/ecp)
[![Go](https://github.com/wrfly/ecp/actions/workflows/go.yml/badge.svg)](https://github.com/wrfly/ecp/actions/workflows/go.yml)
[![GoDoc](https://godoc.org/github.com/wrfly/ecp?status.svg)](https://godoc.org/github.com/wrfly/ecp)
[![license](https://img.shields.io/github/license/wrfly/ecp.svg)](https://github.com/wrfly/ecp/blob/master/LICENSE)

## Usage Example

```go
package main

import (
    "fmt"
    "os"

    "github.com/wrfly/ecp"
)

type Conf struct {
    LogLevel string `default:"debug"`
    Port     int    `env:"PORT"`
}

func main() {
    config := &Conf{}
    if err := ecp.Parse(config); err != nil {
        panic(err)
    }
    fmt.Printf("default log level: [ %s ]\n", config.LogLevel)
    fmt.Println()

    // set some env
    envs := map[string]string{
        "LOGLEVEL": "info",
        "PORT":     "1234",
    }
    for k, v := range envs {
        fmt.Printf("export %s=%s\n", k, v)
        os.Setenv(k, v)
    }
    fmt.Println()

    // then parse configuration from environments
    if err := ecp.Parse(config); err != nil {
        panic(err)
    }
    fmt.Printf("new log level: [ %s ], port: [ %d ]\n",
        config.LogLevel, config.Port)
    fmt.Println()

    // and list all the env keys
    for _, k := range ecp.List(config) {
        fmt.Println(k)
    }
}
```

Outputs:

```txt
default log level: [ debug ]

export LOGLEVEL=info
export PORT=1234

new log level: [ info ], port: [ 1234 ]

LOGLEVEL=debug
PORT=
```

`Parse` needs a pointer to a struct, `ecp.Parse(config)` on a plain struct
returns an error instead of quietly filling in nothing.

## Keys

The key of a field is built from the name of the struct it lives in and the
name of the field itself, upper cased and joined with `_`. A `yaml` or
`json` tag replaces the field name, and an `env` tag replaces the whole key:

```go
type Conf struct {
    LogLevel string                     // LOGLEVEL
    Redis    struct {
        Host string `yaml:"host"`       // REDIS_HOST
        Port int    `env:"REDIS_PORT"`  // REDIS_PORT
    } `yaml:"redis"`
    Secret   string `yaml:"-"`          // ignored
}
```

Pass a prefix to namespace them: `ecp.Parse(&config, "APP")` reads
`APP_LOGLEVEL` and `APP_REDIS_HOST`. A key coming from an `env` tag is
taken as is, so `REDIS_PORT` stays `REDIS_PORT`. The same prefix goes to
`List` and to the `Get` helpers.

## Values

`Parse` sets a field when the environment key exists, otherwise it falls
back to the `default` tag. A default never overwrites a value the caller
already set, an environment value always does.

Supported types are `string`, `bool`, all sized `int`/`uint` types,
`float32`, `float64`, `time.Duration`, slices of those, and named types
built on top of them. Anything else (a `map`, an array, ...) is reported as
an error rather than silently skipped.

- **slices** are separated by a space by default, change it with
  `e := ecp.New(); e.Advance.SplitChar = ","`
- **durations** accept everything `time.ParseDuration` does, plus `Xd`
  for X days: `10s`, `5m`, `6d`
- **integers** also accept `1e3` and `1,000` notation
- **pointers** (`*int`, `*time.Duration`, ...) only get their default when
  they are nil, which makes "unset" and "set to the zero value"
  distinguishable
- **pointers to a struct** are optional sections: they are walked into like
  a plain struct and only allocated when one of their fields is set

An environment variable set to an empty value is treated as unset, so a
field keeps its default.

## Advanced

`ecp.New()` returns a parser whose behaviour can be changed:

```go
e := ecp.New()
e.BuildKey = func(structure, field string, tag reflect.StructTag) string { ... }
e.LookupValue = func(key string) (string, bool) { ... }
e.Advance.SplitChar = ","
e.Advance.SetValue = func(tag reflect.StructTag, field reflect.Value, val string) bool { ... }
```

`LookupValue` is what makes it possible to read from something other than
the environment, and `SetValue` takes over the conversion of a field,
returning true when it handled it.
