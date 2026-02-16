# RTBilling API

This is a modified fork of [UniBee API](https://github.com/UniBee-Billing/unibee-api), a recurring billing system backend built on the [GoFrame](https://github.com/gogf/gf) framework.

**Original project:** [UniBee-Billing/unibee-api](https://github.com/UniBee-Billing/unibee-api)
**Original authors:** [UniBee-Billing](https://github.com/UniBee-Billing)

## License

This project is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**, the same license as the original UniBee API. See the [LICENSE](LICENSE) file for the full text.

As required by AGPL-3.0 Section 13: if you deploy a modified version of this software as a network service, you must make the complete corresponding source code available to all users interacting with it over the network.

## Changes from Upstream

This fork adds SMTP-based email delivery support and other modifications. See the [commit history](../../commits/main) for a detailed record of all changes made from the upstream project.

## Supported Integrations

### Payment Gateways
- Stripe
- PayPal
- Changelly
- Payssion

### VAT
- VatSense

### Email
- SendGrid
- SMTP (added in this fork)

## Requirements

- **MySQL** — `github.com/gogf/gf/contrib/drivers/mysql/v2`
- **Redis** — `github.com/gogf/gf/contrib/nosql/redis/v2`
- **Nacos** (optional) — for configuration management

## Getting Started

### Using Nacos (optional)

Program arguments:
```bash
go run main.go \
  --nacos-ip={your_nacos_ip} \
  --nacos-port={your_nacos_port} \
  --nacos-namespace={your_nacos_namespace} \
  --nacos-group={your_nacos_group} \
  --nacos-data-id={your_nacos_data_id}
```

K8S environment variables:
```
nacos.ip={your_nacos_ip}
nacos.port={your_nacos_port}
nacos.namespace={your_nacos_namespace}
nacos.group={your_nacos_group}
nacos.data.id={your_nacos_data_id}
```

### Without Nacos

Copy `manifest/config/config.yaml.template` to `config.yaml` in the project root, then:
```bash
go run main.go --nacos-enable=false
```

### Local Development

```bash
go run main.go --nacos-port=30099 --nacos-namespace=local --nacos-group=config --nacos-data-id=unib-settings.yaml --nacos-ip=api.unibee.top
```

- OpenAPI V3 Doc: http://127.0.0.1:8088/swagger
- Swagger UI: http://127.0.0.1:8088/swagger-ui.html
- OpenAPI V3 JSON: http://127.0.0.1:8088/api.json

### Code Generation

```bash
gf gen ctrl  # Generate API controllers
gf gen dao   # Generate DAO code after DB table changes
```

See the [GoFrame documentation](https://goframe.org) for more details.

## Project Structure

```
├── api/          # External API input/output definitions
├── hack/         # Dev tools and scripts
├── internal/
│   ├── cmd/      # CLI entry points
│   ├── consts/   # Constants
│   ├── controller/ # Request handling layer
│   ├── dao/      # Data access objects
│   ├── logic/    # Business logic
│   ├── model/
│   │   ├── do/   # Domain objects (generated)
│   │   └── entity/ # Data models (generated)
│   └── service/  # Service interfaces
├── manifest/     # Config, Docker, deploy files
├── resource/     # Static resources
├── utility/      # Utilities
├── go.mod
└── main.go
```
