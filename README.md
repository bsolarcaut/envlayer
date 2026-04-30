# envlayer

A secrets manager shim that resolves environment variables from multiple backends like Vault, SSM, and dotenv at runtime.

---

## Installation

```bash
go install github.com/yourorg/envlayer@latest
```

Or add it as a library:

```bash
go get github.com/yourorg/envlayer
```

## Usage

Define your environment variables with a backend prefix, and `envlayer` will resolve them at startup:

```bash
# .env
DATABASE_URL=ssm:///myapp/prod/database_url
API_KEY=vault://secret/myapp#api_key
REDIS_URL=dotenv://local.env#REDIS_URL
```

Then wrap your application:

```bash
envlayer run -- ./myapp
```

Or use it programmatically:

```go
package main

import (
    "fmt"
    "github.com/yourorg/envlayer"
)

func main() {
    if err := envlayer.Load(); err != nil {
        panic(err)
    }

    // Environment variables are now resolved from their respective backends
    fmt.Println(os.Getenv("DATABASE_URL")) // resolved value from SSM
}
```

### Supported Backends

| Prefix | Backend |
|--------|---------|
| `ssm://` | AWS Systems Manager Parameter Store |
| `vault://` | HashiCorp Vault |
| `dotenv://` | Local `.env` file |

## Configuration

Backend credentials are picked up from standard environment conventions (e.g., `AWS_REGION`, `VAULT_ADDR`, `VAULT_TOKEN`). No extra configuration required.

## License

MIT © yourorg