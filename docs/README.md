# Packer Plugin for IBM Cloud Power Virtual Server - Documentation

Complete documentation for automating custom image creation on IBM Cloud Power Systems infrastructure.

## Table of Contents

- [Installation](#installation)
- [Plugin Components](#plugin-components)
- [Getting Started](#getting-started)
- [Configuration Reference](#configuration-reference)
- [Advanced Usage](#advanced-usage)
- [Best Practices](#best-practices)
- [Additional Resources](#additional-resources)

## Installation

### Using pre-built releases

#### Using the `packer init` command

Starting from version 1.7, Packer supports a new `packer init` command allowing automatic installation of Packer plugins. Read the [Packer documentation](https://www.packer.io/docs/commands/init) for more information.

To install this plugin, copy and paste this code into your Packer configuration. Then, run [`packer init`](https://www.packer.io/docs/commands/init).

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

#### Manual installation

You can find pre-built binary releases of the plugin [here](https://github.com/ppc64le-cloud/packer-plugin-powervs/releases).

Once you have downloaded the latest archive corresponding to your target OS, uncompress it to retrieve the plugin binary file corresponding to your platform.

To install the plugin, please follow the Packer documentation on [installing a plugin](https://www.packer.io/docs/extending/plugins/#installing-plugins).

**Installation Paths:**
- **Linux/macOS**: `~/.packer.d/plugins/`
- **Windows**: `%APPDATA%\packer.d\plugins\`

#### From Source

If you prefer to build the plugin from its source code, clone the GitHub repository locally and run the command `go build` from the root directory:

```bash
git clone https://github.com/ppc64le-cloud/packer-plugin-powervs.git
cd packer-plugin-powervs
make install
```

This will build and install the plugin automatically to your Packer plugins directory.

**For Development:**
```bash
make dev  # Quick build and install for testing
```

## Plugin Components

- **Builder** ([docs](/docs/builders/builder-name.mdx)): Creates custom images from stock images or COS imports
- **Provisioner** ([docs](/docs/provisioners/provisioner-name.mdx)): Configures images during build (shell, Ansible, etc.)
- **Post-Processor** ([docs](/docs/post-processors/postprocessor-name.mdx)): Processes artifacts after build
- **Data Source** ([docs](/docs/datasources/datasource-name.mdx)): Queries PowerVS resources

See [USER_GUIDE.md](USER_GUIDE.md) for detailed component usage.

## Quick Start

For detailed setup instructions, see [USER_GUIDE.md](USER_GUIDE.md).

**Prerequisites:** IBM Cloud account, PowerVS service instance, API key, and SSH key pair.

**Minimal Example:**

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
  api_key             = var.ibm_api_key
  service_instance_id = "your-service-instance-id"
  zone                = "lon04"
  
  source {
    stock_image { name = "CentOS-Stream-8" }
  }
  
  instance_name        = "packer-${timestamp()}"
  key_pair_name        = "my-ssh-key"
  dhcp_network         = true
  ssh_username         = "root"
  ssh_private_key_file = "~/.ssh/id_rsa"
  
  capture {
    name        = "custom-image-${timestamp()}"
    destination = "image-catalog"
  }
}

build {
  sources = ["source.powervs.centos"]

  provisioner "shell" {
    inline = [
      "yum update -y",
      "yum install -y vim wget curl"
    ]
  }
}
```

### Running Your First Build

1. **Create a variables file** (`variables.pkrvars.hcl`):
```hcl
ibm_api_key = "your-api-key-here"
```

2. **Initialize Packer**:
```bash
packer init .
```

3. **Validate the template**:
```bash
packer validate -var-file="variables.pkrvars.hcl" .
```

4. **Build the image**:
```bash
packer build -var-file="variables.pkrvars.hcl" .
```

## Configuration Reference

For complete configuration options, see [API_REFERENCE.md](API_REFERENCE.md).

**Key Configuration Areas:**
- **Authentication**: API key, service instance, zone
- **Source Images**: Stock images or COS imports
- **Instance Settings**: Name, SSH keys, networks
- **Capture Options**: Image catalog, COS, or both
- **SSH Settings**: Username, key file, timeout

## Advanced Usage

### Multi-Stage Builds

Create a base image and then build application-specific images from it:

```hcl
# Stage 1: Base image
source "powervs" "base" {
  # ... configuration ...
  capture {
    name = "base-image-${formatdate("YYYY-MM-DD", timestamp())}"
  }
}

# Stage 2: Application image
source "powervs" "app" {
  source {
    name = "base-image-2024-03-14"  # Reference base image
  }
  # ... configuration ...
  capture {
    name = "app-image-${formatdate("YYYY-MM-DD", timestamp())}"
  }
}

build {
  sources = ["source.powervs.base"]
  # Base provisioning
}

build {
  sources = ["source.powervs.app"]
  # Application-specific provisioning
}
```

### Using Data Sources

Query existing PowerVS resources:

```hcl
data "powervs" "stock_images" {
  # Query available stock images
}

locals {
  latest_centos = data.powervs.stock_images.centos_stream_8
}

source "powervs" "example" {
  source {
    stock_image {
      name = local.latest_centos
    }
  }
  # ... rest of configuration ...
}
```

### Parallel Builds

Build multiple images simultaneously:

```hcl
source "powervs" "centos" {
  # CentOS configuration
}

source "powervs" "rhel" {
  # RHEL configuration
}

build {
  sources = [
    "source.powervs.centos",
    "source.powervs.rhel"
  ]
  
  # Shared provisioning
  provisioner "shell" {
    inline = ["yum update -y"]
  }
}
```

## Best Practices

### Security

1. **Never commit credentials**: Use variables and `.pkrvars.hcl` files (add to `.gitignore`)
2. **Use IAM roles**: When possible, use service IDs with minimal required permissions
3. **Rotate API keys**: Regularly rotate IBM Cloud API keys
4. **Secure SSH keys**: Use strong SSH key pairs and protect private keys

### Performance

1. **Use DHCP networks**: Faster than creating custom networks for temporary builds
2. **Optimize provisioning**: Combine commands to reduce SSH round trips
3. **Parallel builds**: Build multiple images simultaneously when possible
4. **Local caching**: Cache downloaded packages and dependencies

### Reliability

1. **Set appropriate timeouts**: Adjust `cleanup_timeout` based on your needs
2. **Handle errors gracefully**: Use `on_error` in provisioners
3. **Validate templates**: Always run `packer validate` before building
4. **Test incrementally**: Test provisioning scripts separately before full builds

### Cost Optimization

1. **Clean up resources**: Ensure instances are deleted after builds
2. **Use appropriate instance sizes**: Don't over-provision build instances
3. **Minimize build time**: Optimize provisioning to reduce compute costs
4. **Use image catalog**: Avoid COS costs when possible

## Additional Resources

### Documentation

- [Builder Configuration Reference](/docs/builders/builder-name.mdx)
- [Provisioner Guide](/docs/provisioners/provisioner-name.mdx)
- [Post-Processor Reference](/docs/post-processors/postprocessor-name.mdx)
- [Data Source Documentation](/docs/datasources/datasource-name.mdx)
- [Troubleshooting Guide](TROUBLESHOOTING.md)
- [Architecture Overview](ARCHITECTURE.md)

### Examples

- [Apache Web Server](../builder/examples/apache_server/) - Basic web server setup
- [Kubernetes Node Image](image-builder/) - Integration with image-builder
- [Complete Examples](../example/) - Various configuration examples

### External Resources

- [IBM Cloud PowerVS Documentation](https://cloud.ibm.com/docs/power-iaas)
- [Packer Documentation](https://www.packer.io/docs)
- [Packer Plugin SDK](https://github.com/hashicorp/packer-plugin-sdk)
- [IBM Cloud CLI](https://cloud.ibm.com/docs/cli)

## Support

For help and support:

- **Troubleshooting**: See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues and solutions
- **Report Bugs**: [GitHub Issues](https://github.com/ppc64le-cloud/packer-plugin-powervs/issues)
- **Ask Questions**: [GitHub Discussions](https://github.com/ppc64le-cloud/packer-plugin-powervs/discussions)

## Contributing

We welcome contributions! See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on:

- Reporting bugs
- Suggesting features
- Submitting pull requests
- Development setup
- Testing procedures

---

**Last Updated**: March 2024  
**Plugin Version**: 0.0.1  
**Packer Compatibility**: >= 1.7.0
