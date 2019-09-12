# aws-secrets-manager-config

Wraps [caarlos0/env](https://github.com/caarlos0/env) to provide a single view of configuration from the environment and secrets from AWS Secrets Manager.

## Example

```go
package main

import (
	"fmt"
	"time"
	"github.com/JSainsburyPLC/aws-secrets-manager-conf/conf"
)

type config struct {
	ApiKey       string        `secret:"apikey"`
	Port         int           `env:"PORT" envDefault:"3000"`
	IsProduction bool          `env:"PRODUCTION"`
	Hosts        []string      `env:"HOSTS" envSeparator:":"`
	Duration     time.Duration `env:"DURATION"`
}

func main() {
	cfg := config{}
    awsSession  := ...
	conf.Parse(&cfg, "/my-app-secrets/login", secretsmanager.New(awsSession))
}
```

## Constraints

* Secrets must be defined as JSON in AWS Secrets Manager. Keys/Value pairs are mapped to the Go struct based on the `secret` tag.
* Secrets values must be strings
* Secrets are required - they must be defined or the code will panic. We prefer to load the secrets on startup and fail fast if the secret value is not set.