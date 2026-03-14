# Packer Plugin PowerVS - Examples

This directory contains working examples demonstrating various use cases and configurations for the Packer Plugin for IBM Cloud Power Virtual Server.

## Overview

These examples are designed to:
- Demonstrate plugin capabilities
- Provide starting templates for common scenarios
- Serve as integration tests for the plugin
- Help users understand best practices

## Prerequisites

Before running these examples, ensure you have:

1. **IBM Cloud Account**: Active account with PowerVS service
2. **API Key**: IBM Cloud API key with PowerVS permissions
3. **PowerVS Service Instance**: Created in your target zone
4. **SSH Key**: Public key uploaded to PowerVS
5. **Packer**: Version 1.7.0 or higher installed

## Quick Start

### 1. Install the Plugin

```bash
# Initialize Packer and download the plugin
packer init .
```

### 2. Configure Variables

Create a `variables.pkrvars.hcl` file (don't commit this file):

```hcl
ibm_api_key         = "your-ibm-cloud-api-key"
service_instance_id = "your-powervs-service-instance-id"
zone                = "lon04"
ssh_key_name        = "your-ssh-key-name"
ssh_private_key     = "~/.ssh/id_rsa"
```

### 3. Validate the Template

```bash
packer validate -var-file="variables.pkrvars.hcl" .
```

### 4. Build the Image

```bash
packer build -var-file="variables.pkrvars.hcl" .
```

## Example Files

### build.pkr.hcl

Main build configuration demonstrating:
- Plugin initialization
- Source configuration
- Build blocks
- Provisioner usage
- Post-processor integration

**Key Features:**
- Multiple source definitions
- Parallel builds
- Conditional provisioning
- Error handling

### data.pkr.hcl

Data source usage examples:
- Querying PowerVS resources
- Using data in builds
- Dynamic configuration

### variables.pkr.hcl

Variable definitions and local values:
- Input variables
- Local computations
- Variable validation
- Sensitive data handling

## Common Use Cases

### Basic Image Build

Build a simple custom image from a stock image:

```hcl
source "powervs" "basic" {
  api_key            = var.ibm_api_key
  service_instance_id = var.service_instance_id
  zone               = var.zone

  source {
    stock_image {
      name = "CentOS-Stream-8"
    }
  }

  instance_name = "packer-basic-${timestamp()}"
  key_pair_name = var.ssh_key_name
  dhcp_network  = true

  ssh_username         = "root"
  ssh_private_key_file = var.ssh_private_key

  capture {
    name        = "basic-image-${timestamp()}"
    destination = "image-catalog"
  }
}

build {
  sources = ["source.powervs.basic"]

  provisioner "shell" {
    inline = [
      "yum update -y",
      "yum install -y vim wget curl"
    ]
  }
}
```

### Import from Cloud Object Storage

Build from an image stored in COS:

```hcl
source "powervs" "cos_import" {
  api_key            = var.ibm_api_key
  service_instance_id = var.service_instance_id
  zone               = var.zone

  source {
    name = "imported-base-image"
    cos {
      bucket = var.cos_bucket
      object = "base-image.ova.gz"
      region = var.cos_region
    }
  }

  instance_name = "packer-cos-${timestamp()}"
  key_pair_name = var.ssh_key_name
  dhcp_network  = true

  ssh_username         = "root"
  ssh_private_key_file = var.ssh_private_key

  capture {
    name        = "custom-image-${timestamp()}"
    destination = "cloud-storage"
    cos {
      bucket     = var.cos_bucket
      region     = var.cos_region
      access_key = var.cos_access_key
      secret_key = var.cos_secret_key
    }
  }
}
```

### Multi-Stage Build

Create a base image and then build application images from it:

```hcl
# Stage 1: Base image with common tools
source "powervs" "base" {
  # ... configuration ...
  
  capture {
    name        = "base-${formatdate("YYYY-MM-DD", timestamp())}"
    destination = "image-catalog"
  }
}

build {
  name = "base-image"
  sources = ["source.powervs.base"]

  provisioner "shell" {
    inline = [
      "yum update -y",
      "yum install -y vim wget curl git",
      "yum clean all"
    ]
  }
}

# Stage 2: Application image
source "powervs" "app" {
  # ... configuration ...
  
  source {
    name = "base-2024-03-14"  # Reference base image
  }

  capture {
    name        = "app-${formatdate("YYYY-MM-DD", timestamp())}"
    destination = "image-catalog"
  }
}

build {
  name = "app-image"
  sources = ["source.powervs.app"]

  provisioner "shell" {
    inline = [
      "yum install -y httpd",
      "systemctl enable httpd"
    ]
  }
}
```

### Parallel Builds

Build multiple images simultaneously:

```hcl
source "powervs" "centos" {
  # CentOS configuration
  source {
    stock_image {
      name = "CentOS-Stream-8"
    }
  }
  # ... rest of config ...
}

source "powervs" "rhel" {
  # RHEL configuration
  source {
    stock_image {
      name = "RHEL8-SP4"
    }
  }
  # ... rest of config ...
}

build {
  sources = [
    "source.powervs.centos",
    "source.powervs.rhel"
  ]

  # Shared provisioning
  provisioner "shell" {
    inline = [
      "yum update -y",
      "yum install -y common-packages"
    ]
  }

  # OS-specific provisioning
  provisioner "shell" {
    only = ["powervs.centos"]
    inline = ["echo 'CentOS-specific setup'"]
  }

  provisioner "shell" {
    only = ["powervs.rhel"]
    inline = ["echo 'RHEL-specific setup'"]
  }
}
```

### Using Existing Networks

Use existing subnets instead of DHCP:

```hcl
source "powervs" "existing_network" {
  # ... authentication config ...

  subnet_ids = [
    "subnet-id-1",
    "subnet-id-2"
  ]
  dhcp_network = false  # Don't create DHCP network

  # ... rest of config ...
}
```

### With User Data (Cloud-Init)

Provide cloud-init configuration:

```hcl
source "powervs" "cloud_init" {
  # ... configuration ...

  user_data = file("${path.root}/cloud-init.yaml")

  # ... rest of config ...
}
```

**cloud-init.yaml:**
```yaml
#cloud-config
packages:
  - vim
  - wget
  - curl

runcmd:
  - echo "Cloud-init setup complete"
```

## Advanced Examples

### With Ansible Provisioner

```hcl
build {
  sources = ["source.powervs.example"]

  provisioner "ansible" {
    playbook_file = "./playbook.yml"
    user          = "root"
    extra_arguments = [
      "--extra-vars",
      "ansible_python_interpreter=/usr/bin/python3"
    ]
  }
}
```

### With File Uploads

```hcl
build {
  sources = ["source.powervs.example"]

  provisioner "file" {
    source      = "files/config.conf"
    destination = "/etc/myapp/config.conf"
  }

  provisioner "shell" {
    inline = [
      "chmod 644 /etc/myapp/config.conf",
      "systemctl restart myapp"
    ]
  }
}
```

### With Error Handling

```hcl
build {
  sources = ["source.powervs.example"]

  provisioner "shell" {
    inline = [
      "risky-command || true"  # Continue on error
    ]
  }

  provisioner "shell" {
    inline         = ["critical-command"]
    on_error       = "abort"  # Stop on error
    valid_exit_codes = [0, 2]  # Accept these exit codes
  }
}
```

## Testing Examples

### Validate All Examples

```bash
# Validate syntax
packer validate .

# Format templates
packer fmt .

# Check with specific variables
packer validate -var-file="test-variables.pkrvars.hcl" .
```

### Dry Run

```bash
# See what would be built without actually building
packer inspect .
```

### Debug Mode

```bash
# Enable debug logging
export PACKER_LOG=1
export PACKER_LOG_PATH=packer-debug.log

packer build -var-file="variables.pkrvars.hcl" .
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build PowerVS Image

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Setup Packer
        uses: hashicorp/setup-packer@main
        with:
          version: latest
      
      - name: Initialize Packer
        run: packer init .
      
      - name: Validate Template
        run: packer validate .
      
      - name: Build Image
        run: packer build .
        env:
          PKR_VAR_ibm_api_key: ${{ secrets.IBM_API_KEY }}
          PKR_VAR_service_instance_id: ${{ secrets.SERVICE_INSTANCE_ID }}
```

## Best Practices

### 1. Use Variables

```hcl
variable "ibm_api_key" {
  type      = string
  sensitive = true
}

variable "zone" {
  type    = string
  default = "lon04"
}
```

### 2. Use Timestamps for Uniqueness

```hcl
instance_name = "packer-${timestamp()}"
```

### 3. Clean Up Resources

```hcl
cleanup_timeout = "15m"  # Ensure cleanup completes
```

### 4. Use Descriptive Names

```hcl
capture {
  name = "app-v1.2.3-${formatdate("YYYY-MM-DD", timestamp())}"
}
```

### 5. Handle Secrets Properly

```hcl
# Never commit secrets
# Use variables and .pkrvars.hcl (add to .gitignore)
api_key = var.ibm_api_key
```

## Troubleshooting

### Common Issues

1. **Plugin Not Found**
   ```bash
   packer init .  # Download plugin
   ```

2. **Authentication Failed**
   ```bash
   # Verify API key
   ibmcloud login --apikey $IBM_API_KEY
   ```

3. **SSH Timeout**
   ```hcl
   ssh_timeout = "30m"  # Increase timeout
   ```

4. **Build Fails**
   ```bash
   PACKER_LOG=1 packer build .  # Enable debug logging
   ```

## Additional Resources

- [Main Documentation](../docs/README.md)
- [Troubleshooting Guide](../docs/TROUBLESHOOTING.md)
- [Architecture Overview](../docs/ARCHITECTURE.md)
- [Contributing Guide](../CONTRIBUTING.md)

## Support

- **Issues**: [GitHub Issues](https://github.com/ppc64le-cloud/packer-plugin-powervs/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ppc64le-cloud/packer-plugin-powervs/discussions)

---

**Note**: Remember to clean up resources after testing to avoid unnecessary costs.