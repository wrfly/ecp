# ecp

> Environment config parser

If you run your application in a container and deploy it via a docker-compose file, then you may need this tool
for parsing configuration easily instead of mounting an external config file. You can simply set some environments
and then `ecp` will help you fill the configs. Or, you can `COPY` a "default" config file to the image, and change some
variables by overwriting the keys via environments.

The environment config keys can be auto generated or set by the `yaml` or `env` tag.

The only thing you should do is importing this package, and `Parse` your config.

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
    if err := ecp.Default(config); err != nil {
        panic(err)
    }
    fmt.Printf("default log level: [ %s ]\n", config.LogLevel)
    fmt.Println()

    // set some env
    envs := map[string]string{
        "ECP_LOGLEVEL": "info",
        "PORT":         "1234",
    }
    for k, v := range envs {
        fmt.Printf("export %s=%s\n", k, v)
        os.Setenv(k, v)
    }
    fmt.Println()

    if err := ecp.Parse(config, "ECP"); err != nil {
        panic(err)
    }
    fmt.Printf("new log level: [ %s ], port: [ %d ]\n",
        config.LogLevel, config.Port)
    fmt.Println()

    fmt.Println("the env keys are as follows:")
    envLists := ecp.List(config, "ecp")
    for _, k := range envLists {
        fmt.Println(k)
    }
}
```

Outputs:

```txt
default log level: [ debug ]

export ECP_LOGLEVEL=info
export PORT=1234

new log level: [ info ], port: [ 1234 ]

the env keys are as follows:
ECP_LOGLEVEL=debug
PORT=
```