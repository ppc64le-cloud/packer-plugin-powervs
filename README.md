# Packer Plugin for IBM Cloud Power Virtual Server

[![Go Report Card](https://goreportcard.com/badge/github.com/ppc64le-cloud/packer-plugin-powervs)](https://goreportcard.com/report/github.com/ppc64le-cloud/packer-plugin-powervs)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

The Packer Plugin for IBM Cloud Power Virtual Server (PowerVS) enables automated creation of custom images on IBM Cloud's Power Systems infrastructure. This plugin provides a builder, provisioner, post-processor, and data source for managing PowerVS resources through Packer.

## Features

- **Automated Image Building**: Create custom PowerVS images from stock images or Cloud Object Storage (COS) sources
- **Flexible Provisioning**: Support for shell scripts, Ansible, and other Packer provisioners
- **Multiple Capture Options**: Export images to Cloud Object Storage, Image Catalog, or both
- **Network Management**: Automatic DHCP network creation or use existing subnets
- **SSH Support**: Built-in SSH communicator for instance configuration
- **Multi-Architecture**: Native support for ppc64le architecture

## Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Examples](#examples)
- [Documentation](#documentation)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## Requirements

- [Packer](https://www.packer.io/downloads) >= 1.7.0
- [Go](https://golang.org/doc/install) >= 1.18 (for building from source)
- IBM Cloud Account with PowerVS service
- IBM Cloud API Key with appropriate permissions

### IBM Cloud Prerequisites

1. **PowerVS Service Instance**: Create a PowerVS service instance in your desired region
2. **API Key**: Generate an IBM Cloud API key with PowerVS access
3. **SSH Key**: Upload your SSH public key to PowerVS
4. **Network**: Either existing subnet IDs or enable DHCP network creation
5. **Cloud Object Storage** (Optional): For importing/exporting images

## Installation

### Using `packer init` (Recommended)

Add the following to your Packer template:

```hcl
packer {
  required_plugins {
    powervs = {
      version = ">= 0.0.1"
      source  = "github.com/ppc64le-cloud/powervs"
    }
  }
}
```

Then run:

```bash
packer init .
```

### Manual Installation

1. Download the latest release from [GitHub Releases](https://github.com/ppc64le-cloud/packer-plugin-powervs/releases)
2. Extract the binary to your Packer plugins directory:
   - Linux/macOS: `~/.packer.d/plugins/`
   - Windows: `%APPDATA%\packer.d\plugins\`

### Building from Source

```bash
git clone https://github.com/ppc64le-cloud/packer-plugin-powervs.git
cd packer-plugin-powervs
make install
```

This will build the plugin and install it to your Packer plugins directory automatically.

**Alternative: Development Installation**

For development purposes, you can use:

```bash
make dev
```

This builds and copies the binary to `~/.packer.d/plugins/` for quick testing.

## Quick Start

Create a template file `template.pkr.hcl`:

```hcl
packer {
  required_plugins {
    powervs = {
      version = ">= 0.0.1"
      source  = "github.com/ppc64le-cloud/powervs"
    }
  }
}

source "powervs" "example" {
  api_key            = var.ibm_api_key
  service_instance_id = "your-service-instance-id"
  zone               = "lon04"

  source {
    stock_image {
      name = "CentOS-Stream-8"
    }
  }

  instance_name = "packer-${timestamp()}"
  key_pair_name = "your-ssh-key"
  dhcp_network  = true

  ssh_username         = "root"
  ssh_private_key_file = "~/.ssh/id_rsa"

  capture {
    name        = "my-image-${timestamp()}"
    destination = "image-catalog"
  }
}

build {
  sources = ["source.powervs.example"]

  provisioner "shell" {
    inline = ["yum update -y && yum install -y vim"]
  }
}
```

Initialize and build:

```bash
packer init .
packer build -var="ibm_api_key=YOUR_KEY" template.pkr.hcl
```

For detailed configuration options, see the [User Guide](docs/USER_GUIDE.md) and [API Reference](docs/API_REFERENCE.md).

## Examples

- **[Apache Web Server](builder/examples/apache_server)** - Complete example with HTTP server setup
- **[Kubernetes Integration](docs/image-builder)** - Using with Kubernetes Image Builder
- **[More Examples](example/)** - Additional use cases and configurations

## Documentation

- **[User Guide](docs/USER_GUIDE.md)** - Complete guide from basics to advanced usage
- **[API Reference](docs/API_REFERENCE.md)** - Full configuration reference
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[Architecture](docs/ARCHITECTURE.md)** - Plugin design and internals
- **[Plugin Components](docs/README.md)** - Detailed component documentation

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md) for complete development guide including:
- Development setup and workflow
- Building and testing
- Code standards and best practices
- Submitting contributions

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for the complete guide.

Quick start: Fork → Create branch → Make changes → Test → Submit PR

## Troubleshooting

Having issues? Check the [Troubleshooting Guide](docs/TROUBLESHOOTING.md) for solutions to common problems.

## Support

- **Issues**: [GitHub Issues](https://github.com/ppc64le-cloud/packer-plugin-powervs/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ppc64le-cloud/packer-plugin-powervs/discussions)

## Related Projects

- [Packer](https://www.packer.io/) - HashiCorp Packer
- [IBM Cloud Power Virtual Server](https://www.ibm.com/cloud/power-virtual-server)
- [Kubernetes Image Builder](https://github.com/kubernetes-sigs/image-builder)
- [Cluster API Provider IBM Cloud](https://github.com/kubernetes-sigs/cluster-api-provider-ibmcloud)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Maintainers

- [@mkumatag](https://github.com/mkumatag)
- [@Karthik-K-N](https://github.com/Karthik-K-N)

See [OWNERS](OWNERS) for the complete list of maintainers.

## Acknowledgments

- HashiCorp Packer team for the plugin SDK
- IBM Cloud team for PowerVS APIs
- Kubernetes SIG Cluster Lifecycle for image-builder integration
- All contributors who have helped improve this plugin

---

**Note**: This plugin is community-maintained and not officially supported by IBM or HashiCorp.
