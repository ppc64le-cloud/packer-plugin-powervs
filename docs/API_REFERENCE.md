# API Reference

Complete reference documentation for the Packer Plugin for IBM Cloud Power Virtual Server configuration options.

## Table of Contents

- [Builder Configuration](#builder-configuration)
- [Access Configuration](#access-configuration)
- [Source Configuration](#source-configuration)
- [Instance Configuration](#instance-configuration)
- [Network Configuration](#network-configuration)
- [Capture Configuration](#capture-configuration)
- [SSH Configuration](#ssh-configuration)
- [Provisioner Configuration](#provisioner-configuration)
- [Post-Processor Configuration](#post-processor-configuration)
- [Data Source Configuration](#data-source-configuration)

## Builder Configuration

The PowerVS builder creates custom images on IBM Cloud Power Virtual Server.

### Required Plugin Block

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

### Source Block

```hcl
source "powervs" "name" {
  # Configuration options
}
```

## Access Configuration

Authentication and API access configuration.

### Required Fields

#### `api_key` (string)

IBM Cloud API key for authentication.

- **Required**: Yes
- **Type**: String
- **Sensitive**: Yes
- **Example**: `"abc123def456..."`

```hcl
api_key = var.ibm_api_key
```

#### `service_instance_id` (string)

PowerVS service instance identifier.

- **Required**: Yes
- **Type**: String
- **Example**: `"97ff60d4-5b60-4a3d-bb28-34aedc603bf3"`

```hcl
service_instance_id = "97ff60d4-5b60-4a3d-bb28-34aedc603bf3"
```

#### `zone` (string)

PowerVS zone where resources will be created.

- **Required**: Yes
- **Type**: String
- **Valid Values**: `lon04`, `lon06`, `us-south`, `us-east`, `tok04`, `syd04`, `syd05`, `tor01`, `mon01`, `sao01`, `dal10`, `dal12`, `wdc06`, `wdc07`
- **Example**: `"lon04"`

```hcl
zone = "lon04"
```

### Optional Fields

#### `account_id` (string)

IBM Cloud account ID. Auto-detected from API key if not provided.

- **Required**: No
- **Type**: String
- **Default**: Auto-detected
- **Example**: `"a1b2c3d4e5f6..."`

```hcl
account_id = "a1b2c3d4e5f6..."
```

#### `region` (string)

PowerVS region. Auto-detected from zone if not provided.

- **Required**: No
- **Type**: String
- **Default**: Auto-detected from zone
- **Valid Values**: https://cloud.ibm.com/docs/power-iaas?topic=power-iaas-ibm-cloud-reg
- **Example**: `"lon"`

```hcl
region = "lon"
```

#### `debug` (bool)

Enable debug logging for API calls.

- **Required**: No
- **Type**: Boolean
- **Default**: `false`

```hcl
debug = true
```

## Source Configuration

Defines the base image for the build.

### Source Block Structure

```hcl
source {
  name       = "optional-name"
  cos        = { ... }        # Cloud Object Storage source
  stock_image = { ... }       # Stock image source
}
```

### Fields

#### `name` (string)

Name for the imported image (required when using COS).

- **Required**: Conditional (required with `cos`, not used with `stock_image`)
- **Type**: String
- **Example**: `"my-base-image"`

```hcl
source {
  name = "my-base-image"
  cos {
    # ... COS configuration
  }
}
```

#### `cos` (object)

Cloud Object Storage source configuration.

- **Required**: Conditional (either `cos` or `stock_image` required)
- **Type**: Object

**COS Object Fields:**

##### `bucket` (string)

COS bucket name containing the image.

- **Required**: Yes (when using COS)
- **Type**: String
- **Example**: `"my-images-bucket"`

##### `object` (string)

Object key/name of the image file in the bucket.

- **Required**: Yes (when using COS)
- **Type**: String
- **Supported Formats**: `.ova`, `.ova.gz`
- **Example**: `"centos-base.ova.gz"`

##### `region` (string)

COS bucket region.

- **Required**: Yes (when using COS)
- **Type**: String
- **Valid Values**: `us-south`, `us-east`, `eu-gb`, `eu-de`, `jp-tok`, `au-syd`, etc.
- **Example**: `"us-south"`

**Example:**
```hcl
source {
  name = "imported-image"
  cos {
    bucket = "my-images-bucket"
    object = "centos-base.ova.gz"
    region = "us-south"
  }
}
```

#### `stock_image` (object)

Stock image source configuration.

- **Required**: Conditional (either `cos` or `stock_image` required)
- **Type**: Object

**Stock Image Object Fields:**

##### `name` (string)

Name of the stock image to use.

- **Required**: Yes (when using stock_image)
- **Type**: String
- **Available Images**: `CentOS-Stream-8`, `CentOS-Stream-9`, `RHEL8-SP4`, `RHEL8-SP6`, `RHEL9-SP0`, `RHEL9-SP2`, `SLES15-SP3`, `SLES15-SP4`, `Ubuntu-20.04`, `Ubuntu-22.04`
- **Example**: `"CentOS-Stream-8"`

**Example:**
```hcl
source {
  stock_image {
    name = "CentOS-Stream-8"
  }
}
```

## Instance Configuration

Configuration for the temporary build instance.

### Required Fields

#### `instance_name` (string)

Name for the temporary build instance.

- **Required**: Yes
- **Type**: String
- **Must Be**: Unique
- **Recommendation**: Include timestamp for uniqueness
- **Example**: `"packer-build-${timestamp()}"`

```hcl
instance_name = "packer-build-${timestamp()}"
```

#### `key_pair_name` (string)

Name of the SSH key pair in PowerVS.

- **Required**: Yes
- **Type**: String
- **Example**: `"my-ssh-key"`

```hcl
key_pair_name = "my-ssh-key"
```

### Optional Fields

#### `user_data` (string)

Cloud-init user data for instance initialization.

- **Required**: No
- **Type**: String
- **Format**: YAML or shell script
- **Example**: `file("${path.root}/cloud-init.yaml")`

```hcl
user_data = file("${path.root}/cloud-init.yaml")
```

#### `cleanup_timeout` (string)

Maximum time to wait for instance deletion during cleanup.

- **Required**: No
- **Type**: Duration string
- **Default**: `"10m"`
- **Format**: `"10m"`, `"15m30s"`, `"1h"`
- **Example**: `"15m"`

```hcl
cleanup_timeout = "15m"
```

## Network Configuration

Network configuration for the build instance.

### Option 1: DHCP Network

#### `dhcp_network` (bool)

Automatically create a DHCP network for the build.

- **Required**: Conditional (either `dhcp_network` or `subnet_ids` required)
- **Type**: Boolean
- **Default**: `false`
- **Recommendation**: Use `true` for faster builds

```hcl
dhcp_network = true
```

### Option 2: Existing Subnets

#### `subnet_ids` (list of strings)

List of existing subnet IDs to attach to the instance.

- **Required**: Conditional (either `dhcp_network` or `subnet_ids` required)
- **Type**: List of strings
- **Example**: `["subnet-id-1", "subnet-id-2"]`

```hcl
subnet_ids = [
  "subnet-abc123",
  "subnet-def456"
]
```

## Capture Configuration

Configuration for capturing the built image.

### Capture Block Structure

```hcl
capture {
  name        = "image-name"
  destination = "cloud-storage"  # or "image-catalog" or "both"
  cos         = { ... }           # Required for cloud-storage
}
```

### Required Fields

#### `name` (string)

Name for the captured image.

- **Required**: Yes
- **Type**: String
- **Recommendation**: Include version and timestamp
- **Example**: `"app-v1.0-${timestamp()}"`

```hcl
capture {
  name = "my-image-${timestamp()}"
}
```

### Optional Fields

#### `destination` (string)

Where to save the captured image.

- **Required**: No
- **Type**: String
- **Default**: `"cloud-storage"`
- **Valid Values**: 
  - `"cloud-storage"`: Export to Cloud Object Storage only
  - `"image-catalog"`: Save to PowerVS image catalog only
  - `"both"`: Export to both COS and image catalog
- **Example**: `"both"`

```hcl
capture {
  name        = "my-image"
  destination = "both"
}
```

#### `cos` (object)

Cloud Object Storage destination configuration.

- **Required**: Conditional (required when `destination` is `"cloud-storage"` or `"both"`)
- **Type**: Object

**COS Object Fields:**

##### `bucket` (string)

Destination COS bucket name.

- **Required**: Yes (when using COS destination)
- **Type**: String
- **Example**: `"my-images-bucket"`

##### `region` (string)

COS bucket region.

- **Required**: Yes (when using COS destination)
- **Type**: String
- **Example**: `"us-south"`

##### `access_key` (string)

COS HMAC access key.

- **Required**: Yes (when using COS destination)
- **Type**: String
- **Sensitive**: Yes
- **Example**: `var.cos_access_key`

##### `secret_key` (string)

COS HMAC secret key.

- **Required**: Yes (when using COS destination)
- **Type**: String
- **Sensitive**: Yes
- **Example**: `var.cos_secret_key`

**Example:**
```hcl
capture {
  name        = "my-image-${timestamp()}"
  destination = "cloud-storage"
  cos {
    bucket     = "my-images-bucket"
    region     = "us-south"
    access_key = var.cos_access_key
    secret_key = var.cos_secret_key
  }
}
```

## SSH Configuration

SSH communicator configuration for connecting to the build instance.

### Required Fields

#### `ssh_username` (string)

SSH username for connecting to the instance.

- **Required**: Yes
- **Type**: String
- **Common Values**: `"root"` (Linux), `"ubuntu"` (Ubuntu), `"centos"` (CentOS)
- **Example**: `"root"`

```hcl
ssh_username = "root"
```

#### `ssh_private_key_file` (string)

Path to SSH private key file.

- **Required**: Yes (unless using `ssh_password`)
- **Type**: String
- **Example**: `"~/.ssh/id_rsa"`

```hcl
ssh_private_key_file = "~/.ssh/id_rsa"
```

### Optional Fields

#### `ssh_timeout` (string)

Maximum time to wait for SSH connection.

- **Required**: No
- **Type**: Duration string
- **Default**: `"20m"`
- **Example**: `"30m"`

```hcl
ssh_timeout = "30m"
```

#### `ssh_port` (int)

SSH port number.

- **Required**: No
- **Type**: Integer
- **Default**: `22`
- **Example**: `22`

```hcl
ssh_port = 22
```

#### `ssh_password` (string)

SSH password (alternative to private key).

- **Required**: No
- **Type**: String
- **Sensitive**: Yes
- **Note**: Private key authentication is recommended

```hcl
ssh_password = var.ssh_password
```

## Provisioner Configuration

Provisioners configure the instance after it's created.

### Shell Provisioner

Execute shell commands or scripts.

```hcl
provisioner "shell" {
  inline = [
    "yum update -y",
    "yum install -y httpd"
  ]
}
```

**Options:**

- `inline` (list): Commands to execute
- `script` (string): Path to script file
- `scripts` (list): List of script files
- `environment_vars` (list): Environment variables
- `execute_command` (string): Command wrapper
- `on_error` (string): Error handling (`continue`, `abort`, `ask`)
- `valid_exit_codes` (list): Acceptable exit codes

### File Provisioner

Upload files to the instance.

```hcl
provisioner "file" {
  source      = "local/path/file.conf"
  destination = "/tmp/file.conf"
}
```

**Options:**

- `source` (string): Local file path
- `destination` (string): Remote file path
- `direction` (string): `upload` or `download`

### Ansible Provisioner

Run Ansible playbooks.

```hcl
provisioner "ansible" {
  playbook_file = "playbook.yml"
  user          = "root"
}
```

**Options:**

- `playbook_file` (string): Path to playbook
- `user` (string): SSH user
- `extra_arguments` (list): Additional ansible-playbook arguments
- `ansible_env_vars` (list): Environment variables

## Post-Processor Configuration

Post-processors handle artifacts after the build.

```hcl
post-processor "powervs" {
  # Post-processor configuration
}
```

## Data Source Configuration

Data sources query PowerVS resources.

```hcl
data "powervs" "images" {
  # Data source configuration
}
```

## Complete Example

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

variable "cos_access_key" {
  type      = string
  sensitive = true
}

variable "cos_secret_key" {
  type      = string
  sensitive = true
}

source "powervs" "example" {
  # Access Configuration
  api_key            = var.ibm_api_key
  service_instance_id = "97ff60d4-5b60-4a3d-bb28-34aedc603bf3"
  zone               = "lon04"
  debug              = false

  # Source Configuration
  source {
    stock_image {
      name = "CentOS-Stream-8"
    }
  }

  # Instance Configuration
  instance_name   = "packer-${timestamp()}"
  key_pair_name   = "my-ssh-key"
  cleanup_timeout = "15m"

  # Network Configuration
  dhcp_network = true

  # SSH Configuration
  ssh_username         = "root"
  ssh_private_key_file = "~/.ssh/id_rsa"
  ssh_timeout          = "20m"

  # Capture Configuration
  capture {
    name        = "custom-image-${timestamp()}"
    destination = "both"
    cos {
      bucket     = "my-images-bucket"
      region     = "us-south"
      access_key = var.cos_access_key
      secret_key = var.cos_secret_key
    }
  }
}

build {
  sources = ["source.powervs.example"]

  provisioner "shell" {
    inline = [
      "yum update -y",
      "yum install -y vim wget curl"
    ]
  }

  provisioner "file" {
    source      = "files/config.conf"
    destination = "/tmp/config.conf"
  }

  provisioner "shell" {
    inline = [
      "sudo mv /tmp/config.conf /etc/app/config.conf"
    ]
  }
}
```

## Field Summary Tables

### Access Configuration Summary

| Field | Required | Type | Default | Description |
|-------|----------|------|---------|-------------|
| `api_key` | Yes | string | - | IBM Cloud API key |
| `service_instance_id` | Yes | string | - | PowerVS service instance ID |
| `zone` | Yes | string | - | PowerVS zone |
| `account_id` | No | string | Auto-detected | IBM Cloud account ID |
| `region` | No | string | Auto-detected | PowerVS region |
| `debug` | No | bool | `false` | Enable debug logging |

### Source Configuration Summary

| Field | Required | Type | Default | Description |
|-------|----------|------|---------|-------------|
| `name` | Conditional | string | - | Image name (for COS import) |
| `cos.bucket` | Conditional | string | - | COS bucket name |
| `cos.object` | Conditional | string | - | Image file name |
| `cos.region` | Conditional | string | - | COS region |
| `stock_image.name` | Conditional | string | - | Stock image name |

### Instance Configuration Summary

| Field | Required | Type | Default | Description |
|-------|----------|------|---------|-------------|
| `instance_name` | Yes | string | - | Build instance name |
| `key_pair_name` | Yes | string | - | SSH key pair name |
| `user_data` | No | string | - | Cloud-init user data |
| `cleanup_timeout` | No | string | `"10m"` | Cleanup timeout |

### Network Configuration Summary

| Field | Required | Type | Default | Description |
|-------|----------|------|---------|-------------|
| `dhcp_network` | Conditional | bool | `false` | Create DHCP network |
| `subnet_ids` | Conditional | list | - | Existing subnet IDs |

### Capture Configuration Summary

| Field | Required | Type | Default | Description |
|-------|----------|------|---------|-------------|
| `name` | Yes | string | - | Captured image name |
| `destination` | No | string | `"cloud-storage"` | Capture destination |
| `cos.bucket` | Conditional | string | - | COS bucket name |
| `cos.region` | Conditional | string | - | COS region |
| `cos.access_key` | Conditional | string | - | COS access key |
| `cos.secret_key` | Conditional | string | - | COS secret key |

### SSH Configuration Summary

| Field | Required | Type | Default | Description |
|-------|----------|------|---------|-------------|
| `ssh_username` | Yes | string | - | SSH username |
| `ssh_private_key_file` | Yes | string | - | Private key path |
| `ssh_timeout` | No | string | `"20m"` | SSH timeout |
| `ssh_port` | No | int | `22` | SSH port |

---

**Last Updated**: March 2024  
**Plugin Version**: 0.0.1  
**Packer Compatibility**: >= 1.7.0