# User Guide - Packer Plugin for IBM Cloud Power Virtual Server

This comprehensive guide walks you through using the Packer Plugin for IBM Cloud Power Virtual Server to create custom images.

## Table of Contents

- [Introduction](#introduction)
- [Getting Started](#getting-started)
- [Configuration Guide](#configuration-guide)
- [Building Images](#building-images)
- [Advanced Topics](#advanced-topics)
- [Best Practices](#best-practices)
- [FAQ](#faq)

## Introduction

### What is This Plugin?

The Packer Plugin for PowerVS automates the creation of custom virtual machine images on IBM Cloud's Power Systems infrastructure. It integrates with HashiCorp Packer to provide a declarative way to build, provision, and capture PowerVS images.

### Why Use This Plugin?

- **Automation**: Eliminate manual image creation steps
- **Consistency**: Ensure identical images across environments
- **Version Control**: Track image configurations in Git
- **CI/CD Integration**: Automate image builds in pipelines
- **Multi-Architecture**: Native ppc64le support

### Use Cases

1. **Base Image Creation**: Build standardized base images for your organization
2. **Application Deployment**: Pre-install applications in images
3. **Kubernetes Nodes**: Create custom node images for Kubernetes clusters
4. **Development Environments**: Build consistent dev/test environments
5. **Compliance**: Ensure images meet security and compliance requirements

## Getting Started

### Prerequisites Checklist

Before you begin, ensure you have:

- [ ] IBM Cloud account with PowerVS service enabled
- [ ] PowerVS service instance created in your target zone
- [ ] IBM Cloud API key with PowerVS permissions
- [ ] SSH key pair generated and public key uploaded to PowerVS
- [ ] Packer 1.7.0 or higher installed
- [ ] (Optional) Cloud Object Storage bucket for image import/export

### Step 1: Install Packer

**macOS:**
```bash
brew tap hashicorp/tap
brew install hashicorp/tap/packer
```

**Linux:**
```bash
curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -
sudo apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"
sudo apt-get update && sudo apt-get install packer
```

**Windows:**
```powershell
choco install packer
```

**Verify Installation:**
```bash
packer version
# Should show version >= 1.7.0
```

### Step 2: Install Plugin from Source (Optional)

If you want to build from source instead of using `packer init`:

```bash
git clone https://github.com/ppc64le-cloud/packer-plugin-powervs.git
cd packer-plugin-powervs
make install
```

This builds and installs the plugin to your Packer plugins directory.

**For Development:**
```bash
make dev  # Quick build and install for testing
```

### Step 3: Set Up IBM Cloud

#### Create API Key

1. Log in to [IBM Cloud Console](https://cloud.ibm.com)
2. Navigate to **Manage** → **Access (IAM)** → **API keys**
3. Click **Create an IBM Cloud API key**
4. Give it a descriptive name (e.g., "packer-powervs")
5. Click **Create** and save the API key securely

#### Create PowerVS Service Instance

1. Go to **Catalog** → **Power Systems Virtual Server**
2. Select your region and zone
3. Choose a service plan
4. Create the service instance
5. Note the **Service Instance ID** (found in service details)

#### Upload SSH Key

1. In PowerVS service instance, go to **SSH keys**
2. Click **Create SSH key**
3. Paste your public key content
4. Give it a name and save

### Step 4: Create Your First Template

Create a file named `template.pkr.hcl`:

```hcl
packer {
  required_plugins {
    powervs = {
      version = ">= 0.0.1"
      source  = "github.com/ppc64le-cloud/powervs"
    }
  }
}

variable "ibm_api_key" {
  type      = string
  sensitive = true
}

variable "service_instance_id" {
  type = string
}

variable "zone" {
  type    = string
  default = "lon04"
}

variable "ssh_key_name" {
  type = string
}

source "powervs" "example" {
  # Authentication
  api_key            = var.ibm_api_key
  service_instance_id = var.service_instance_id
  zone               = var.zone

  # Source Image
  source {
    stock_image {
      name = "CentOS-Stream-8"
    }
  }

  # Instance Configuration
  instance_name = "packer-${timestamp()}"
  key_pair_name = var.ssh_key_name
  dhcp_network  = true

  # SSH Configuration
  ssh_username         = "root"
  ssh_private_key_file = "~/.ssh/id_rsa"

  # Capture Configuration
  capture {
    name        = "my-first-image-${timestamp()}"
    destination = "image-catalog"
  }
}

build {
  sources = ["source.powervs.example"]

  provisioner "shell" {
    inline = [
      "echo 'Hello from Packer!'",
      "yum update -y",
      "yum install -y vim wget curl"
    ]
  }
}
```

### Step 5: Create Variables File

Create `variables.pkrvars.hcl` (add to `.gitignore`):

```hcl
ibm_api_key         = "your-api-key-here"
service_instance_id = "your-service-instance-id"
zone                = "lon04"
ssh_key_name        = "your-ssh-key-name"
```

### Step 6: Initialize and Build

```bash
# Initialize Packer (downloads plugin)
packer init .

# Validate template
packer validate -var-file="variables.pkrvars.hcl" .

# Build image
packer build -var-file="variables.pkrvars.hcl" .
```

## Configuration Guide

### Authentication Configuration

#### Using API Key (Recommended)

```hcl
source "powervs" "example" {
  api_key = var.ibm_api_key
  # ... other config
}
```

#### Account Configuration

```hcl
source "powervs" "example" {
  api_key   = var.ibm_api_key
  account_id = "your-account-id"  # Optional, auto-detected if not provided
  # ... other config
}
```

#### Debug Mode

```hcl
source "powervs" "example" {
  debug = true  # Enable detailed logging
  # ... other config
}
```

### Source Image Configuration

#### Option 1: Stock Images

Use pre-built images from IBM:

```hcl
source {
  stock_image {
    name = "CentOS-Stream-8"
  }
}
```

**Available Stock Images:**

PowerVS contains the stock images like CentOS, RHEL, SLES, AIX etc..

**List available images:**
```bash
ibmcloud pi images --instance-id YOUR_SERVICE_INSTANCE_ID
```

#### Option 2: Cloud Object Storage

Import custom images from COS:

```hcl
source {
  name = "my-custom-base"
  cos {
    bucket = "my-cos-bucket"
    object = "centos-custom.ova.gz"
    region = "us-south"
  }
}
```

**Supported Formats:**
- OVA (Open Virtualization Archive)

### Network Configuration

#### Option 1: DHCP Network (Recommended)

Automatically create a DHCP network:

```hcl
dhcp_network = true
```

**Advantages:**
- Fastest setup
- Automatic IP assignment
- No manual network configuration needed

#### Option 2: Existing Subnets

Use existing network subnets:

```hcl
subnet_ids = [
  "subnet-id-1",
  "subnet-id-2"
]
dhcp_network = false
```

**List available subnets:**
```bash
ibmcloud pi networks --instance-id YOUR_SERVICE_INSTANCE_ID
```

### Instance Configuration

```hcl
# Instance name (must be unique)
instance_name = "packer-build-${timestamp()}"

# SSH key for access
key_pair_name = "my-ssh-key"

# Optional: Cloud-init user data
user_data = file("${path.root}/cloud-init.yaml")

# Optional: Cleanup timeout
cleanup_timeout = "15m"
```

### SSH Configuration

```hcl
# SSH username (usually 'root' for Linux)
ssh_username = "root"

# Path to private key
ssh_private_key_file = "~/.ssh/id_rsa"

# Optional: SSH timeout
ssh_timeout = "20m"

# Optional: SSH port
ssh_port = 22
```

### Capture Configuration

#### To Image Catalog Only

```hcl
capture {
  name        = "my-image-${timestamp()}"
  destination = "image-catalog"
}
```

#### To Cloud Object Storage Only

```hcl
capture {
  name        = "my-image-${timestamp()}"
  destination = "cloud-storage"
  cos {
    bucket     = "my-bucket"
    region     = "us-south"
    access_key = var.cos_access_key
    secret_key = var.cos_secret_key
  }
}
```

#### To Both

```hcl
capture {
  name        = "my-image-${timestamp()}"
  destination = "both"
  cos {
    bucket     = "my-bucket"
    region     = "us-south"
    access_key = var.cos_access_key
    secret_key = var.cos_secret_key
  }
}
```

## Building Images

### Basic Build

```bash
packer build template.pkr.hcl
```

### With Variables

```bash
# From file
packer build -var-file="variables.pkrvars.hcl" template.pkr.hcl

# From command line
packer build \
  -var "ibm_api_key=YOUR_KEY" \
  -var "service_instance_id=YOUR_ID" \
  template.pkr.hcl

# From environment
export PKR_VAR_ibm_api_key="YOUR_KEY"
packer build template.pkr.hcl
```

### Debug Build

```bash
# Enable debug logging
export PACKER_LOG=1
export PACKER_LOG_PATH=packer.log

packer build template.pkr.hcl
```

### Parallel Builds

Build multiple images simultaneously:

```hcl
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

### Selective Builds

Build only specific sources:

```bash
# Build only centos
packer build -only="powervs.centos" template.pkr.hcl

# Build except rhel
packer build -except="powervs.rhel" template.pkr.hcl
```

## Advanced Topics

### Multi-Stage Builds

Create a base image, then build specialized images from it:

```hcl
# Stage 1: Base image
source "powervs" "base" {
  source {
    stock_image {
      name = "CentOS-Stream-8"
    }
  }
  capture {
    name = "base-${formatdate("YYYY-MM-DD", timestamp())}"
  }
}

build {
  name = "base"
  sources = ["source.powervs.base"]
  
  provisioner "shell" {
    inline = [
      "yum update -y",
      "yum install -y common-tools"
    ]
  }
}

# Stage 2: Application image
source "powervs" "app" {
  source {
    name = "base-2024-03-14"  # Reference base image
  }
  capture {
    name = "app-${formatdate("YYYY-MM-DD", timestamp())}"
  }
}

build {
  name = "app"
  sources = ["source.powervs.app"]
  
  provisioner "shell" {
    inline = [
      "yum install -y application-packages"
    ]
  }
}
```

### Using Provisioners

#### Shell Provisioner

```hcl
provisioner "shell" {
  inline = [
    "yum update -y",
    "yum install -y httpd"
  ]
}

# Or from file
provisioner "shell" {
  script = "scripts/setup.sh"
}

# Or multiple scripts
provisioner "shell" {
  scripts = [
    "scripts/base.sh",
    "scripts/app.sh"
  ]
}
```

#### File Provisioner

```hcl
provisioner "file" {
  source      = "files/config.conf"
  destination = "/tmp/config.conf"
}

provisioner "shell" {
  inline = [
    "sudo mv /tmp/config.conf /etc/app/config.conf"
  ]
}
```

#### Ansible Provisioner

```hcl
provisioner "ansible" {
  playbook_file = "playbook.yml"
  user          = "root"
  extra_arguments = [
    "--extra-vars",
    "env=production"
  ]
}
```

### Using Data Sources

Query PowerVS resources:

```hcl
data "powervs" "images" {
  # Query configuration
}

locals {
  latest_image = data.powervs.images.centos_latest
}

source "powervs" "example" {
  source {
    name = local.latest_image
  }
}
```

### Template Functions

Use built-in functions:

```hcl
# Timestamp
instance_name = "packer-${timestamp()}"

# Date formatting
image_name = "image-${formatdate("YYYY-MM-DD", timestamp())}"

# UUID
unique_id = "${uuidv4()}"

# File reading
user_data = file("${path.root}/cloud-init.yaml")

# Environment variables
api_key = env("IBM_API_KEY")
```

### Conditional Logic

```hcl
variable "environment" {
  type = string
}

locals {
  instance_type = var.environment == "prod" ? "large" : "small"
}

provisioner "shell" {
  only = ["powervs.production"]
  inline = ["echo 'Production setup'"]
}
```

## Best Practices

### 1. Security

```hcl
# Use variables for secrets
variable "ibm_api_key" {
  type      = string
  sensitive = true  # Prevents logging
}

# Never commit credentials
# Add to .gitignore:
# *.pkrvars.hcl
# packer.log
```

### 2. Naming Conventions

```hcl
# Include timestamps for uniqueness
instance_name = "packer-${timestamp()}"

# Use descriptive names
capture {
  name = "app-v1.2.3-${formatdate("YYYY-MM-DD", timestamp())}"
}
```

### 3. Error Handling

```hcl
provisioner "shell" {
  inline = [
    "command || true"  # Continue on error
  ]
}

provisioner "shell" {
  inline         = ["critical-command"]
  on_error       = "abort"  # Stop on error
  valid_exit_codes = [0, 2]
}
```

### 4. Resource Cleanup

```hcl
# Set appropriate timeout
cleanup_timeout = "15m"

# Verify cleanup in logs
# Check for "Cleanup complete" messages
```

### 5. Testing

```bash
# Validate before building
packer validate .

# Format templates
packer fmt .

# Use debug mode for troubleshooting
PACKER_LOG=1 packer build .
```

### 6. Version Control

```bash
# Track templates in Git
git add template.pkr.hcl
git commit -m "Add web server image template"

# Tag releases
git tag -a v1.0.0 -m "Release version 1.0.0"
```

### 7. Documentation

```hcl
# Add comments to templates
# This builds a web server image with Apache
source "powervs" "webserver" {
  # ... configuration
}

# Document variables
variable "zone" {
  type        = string
  description = "PowerVS zone (e.g., lon04, us-south)"
  default     = "lon04"
}
```

## FAQ

### General Questions

**Q: How long does a build take?**  
A: Typically 45-60 minutes, depending on provisioning complexity and network speed.

**Q: Can I build multiple images in parallel?**  
A: Yes, define multiple sources in a single build block.

**Q: What happens if a build fails?**  
A: Packer automatically cleans up temporary resources (instances, networks).

**Q: Can I resume a failed build?**  
A: No, you need to restart the build. Consider breaking complex builds into stages.

### Configuration Questions

**Q: Which source image should I use?**  
A: Start with stock images for simplicity. Use COS for custom base images.

**Q: Should I use DHCP or existing subnets?**  
A: Use DHCP for faster builds. Use existing subnets if you need specific network configuration.

**Q: Where should I capture images?**  
A: Use image-catalog for quick access. Use cloud-storage for backup and distribution.

### Troubleshooting Questions

**Q: Build times out waiting for SSH?**  
A: Increase `ssh_timeout`, verify SSH key, check network configuration.

**Q: Instance creation fails?**  
A: Check quota limits, verify zone availability, ensure unique instance names.

**Q: Image capture fails?**  
A: Verify COS credentials, check bucket permissions, ensure sufficient storage.

**Q: How do I debug build issues?**  
A: Enable debug logging with `PACKER_LOG=1` and check `packer.log`.

### Cost Questions

**Q: How much does building cost?**  
A: Costs include instance runtime, storage, and data transfer. Minimize build time to reduce costs.

## Next Steps

- Review [Examples](../example/README.md) for more use cases
- Check [Troubleshooting Guide](TROUBLESHOOTING.md) for common issues
- Read [Architecture](ARCHITECTURE.md) to understand internals
- See [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for help and support options

---

**Happy Building!** 🚀