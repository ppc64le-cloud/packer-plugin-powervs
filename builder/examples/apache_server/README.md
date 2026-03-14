# Apache Web Server Example

This example demonstrates how to build a custom PowerVS image with Apache HTTP Server installed and configured.

## Overview

This example:
- Imports a CentOS Stream 8 base image from Cloud Object Storage
- Creates a temporary PowerVS instance
- Installs and configures Apache HTTP Server
- Captures the customized image to Cloud Object Storage

## Prerequisites

1. **IBM Cloud Account** with PowerVS service
2. **PowerVS Service Instance** in your target zone
3. **IBM Cloud API Key** with PowerVS permissions
4. **SSH Key Pair** uploaded to PowerVS
5. **Cloud Object Storage (COS)**:
   - Bucket for source image
   - Bucket for destination image (can be the same)
   - HMAC credentials (access key and secret key)
6. **Base Image** in COS (e.g., `capibm-powervs-centos-streams8-1-22-4.ova.gz`)

## Files

- **apache.json**: Legacy JSON template (Packer < 1.7)
- **apache-stock.json**: Simplified version using stock images
- **packages.sh**: Shell script to install and configure Apache
- **README.md**: This file

## Quick Start

### Option 1: Using Stock Image (Recommended)

This is the simplest approach, using a stock CentOS image directly from PowerVS:

```bash
# Build using stock image
packer build \
  -var "apikey=YOUR_API_KEY" \
  -var "service_instance_id=YOUR_SERVICE_INSTANCE_ID" \
  -var "zone=lon04" \
  -var "ssh_key_name=YOUR_SSH_KEY" \
  -var "access_key=YOUR_COS_ACCESS_KEY" \
  -var "secret_key=YOUR_COS_SECRET_KEY" \
  ./apache-stock.json
```

### Option 2: Using COS Image

If you have a custom base image in Cloud Object Storage:

```bash
# Build using COS image
packer build \
  -var "apikey=YOUR_API_KEY" \
  -var "access_key=YOUR_COS_ACCESS_KEY" \
  -var "secret_key=YOUR_COS_SECRET_KEY" \
  ./apache.json
```

## Configuration

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `apikey` | IBM Cloud API key | `abc123...` |
| `access_key` | COS HMAC access key | `xyz789...` |
| `secret_key` | COS HMAC secret key | `secret123...` |

### Optional Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `shellpath` | Path to provisioning script | `packages.sh` |
| `service_instance_id` | PowerVS service instance ID | (from template) |
| `zone` | PowerVS zone | `lon04` |
| `ssh_key_name` | SSH key name in PowerVS | (from template) |

### Customizing the Template

Edit `apache.json` or `apache-stock.json` to customize:

1. **Service Instance ID**:
   ```json
   "service_instance_id": "YOUR-SERVICE-INSTANCE-ID"
   ```

2. **Zone**:
   ```json
   "zone": "lon04"  // or us-south, us-east, etc.
   ```

3. **SSH Key**:
   ```json
   "key_pair_name": "YOUR-SSH-KEY-NAME"
   ```

4. **SSH Private Key Path**:
   ```json
   "ssh_private_key_file": "/path/to/your/private/key"
   ```

5. **Source Image** (for COS import):
   ```json
   "source": {
     "name": "my-base-image",
     "cos": {
       "bucket": "your-bucket",
       "object": "your-image.ova.gz",
       "region": "us-south"
     }
   }
   ```

6. **Capture Configuration**:
   ```json
   "capture": {
     "name": "apache-server-image",
     "cos": {
       "bucket": "your-destination-bucket",
       "region": "us-south",
       "access_key": "{{user `access_key`}}",
       "secret_key": "{{user `secret_key`}}"
     }
   }
   ```

## Provisioning Script

The `packages.sh` script performs the following:

```bash
#!/bin/bash
set -e

echo "Installing Apache HTTP Server"
sudo yum update -y
sudo yum install -y httpd

echo "Configuring Apache"
sudo systemctl enable httpd
sudo systemctl start httpd

echo "Apache installation complete"
```

### Customizing the Script

You can modify `packages.sh` to:

1. **Install Additional Packages**:
   ```bash
   sudo yum install -y httpd mod_ssl php
   ```

2. **Configure Apache**:
   ```bash
   sudo sed -i 's/Listen 80/Listen 8080/' /etc/httpd/conf/httpd.conf
   ```

3. **Add Custom Content**:
   ```bash
   echo "<h1>Welcome</h1>" | sudo tee /var/www/html/index.html
   ```

4. **Configure Firewall**:
   ```bash
   sudo firewall-cmd --permanent --add-service=http
   sudo firewall-cmd --reload
   ```

## Step-by-Step Build Process

### 1. Prepare Environment

```bash
# Set environment variables (optional)
export PKR_VAR_apikey="your-api-key"
export PKR_VAR_access_key="your-cos-access-key"
export PKR_VAR_secret_key="your-cos-secret-key"
```

### 2. Validate Template

```bash
packer validate apache-stock.json
```

### 3. Build Image

```bash
# With environment variables
packer build apache-stock.json

# Or with command-line variables
packer build \
  -var "apikey=YOUR_API_KEY" \
  -var "access_key=YOUR_COS_ACCESS_KEY" \
  -var "secret_key=YOUR_COS_SECRET_KEY" \
  apache-stock.json
```

### 4. Monitor Progress

The build process will:
1. Import or use the base image
2. Create a temporary instance
3. Wait for SSH connectivity
4. Run the provisioning script
5. Capture the instance as an image
6. Export to Cloud Object Storage
7. Clean up temporary resources

### 5. Verify Image

After the build completes:

```bash
# List images in PowerVS
ibmcloud pi images --instance-id YOUR_SERVICE_INSTANCE_ID

# Or check COS bucket
aws s3 ls s3://your-bucket/ --endpoint-url=https://s3.us-south.cloud-object-storage.appdomain.cloud
```

## Using the Built Image

### Deploy Instance from Image

```bash
# Using IBM Cloud CLI
ibmcloud pi instance-create \
  --name apache-server \
  --image apache-server-image \
  --subnets SUBNET_ID \
  --key-name YOUR_SSH_KEY \
  --instance-id YOUR_SERVICE_INSTANCE_ID
```

### Test Apache

```bash
# SSH to the instance
ssh root@INSTANCE_IP

# Check Apache status
systemctl status httpd

# Test web server
curl http://localhost
```

## Advanced Customization

### Multi-Stage Build

Build a base image first, then build Apache on top:

```json
{
  "builders": [{
    "type": "powervs",
    "source": {
      "name": "base-image-2024-03-14"
    },
    "capture": {
      "name": "apache-server-{{timestamp}}"
    }
  }]
}
```

### With SSL/TLS

Modify `packages.sh`:

```bash
# Install mod_ssl
sudo yum install -y httpd mod_ssl

# Generate self-signed certificate
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/pki/tls/private/apache-selfsigned.key \
  -out /etc/pki/tls/certs/apache-selfsigned.crt \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"

# Configure SSL
sudo sed -i 's/SSLCertificateFile.*/SSLCertificateFile \/etc\/pki\/tls\/certs\/apache-selfsigned.crt/' /etc/httpd/conf.d/ssl.conf
sudo sed -i 's/SSLCertificateKeyFile.*/SSLCertificateKeyFile \/etc\/pki\/tls\/private\/apache-selfsigned.key/' /etc/httpd/conf.d/ssl.conf

# Enable and start
sudo systemctl enable httpd
sudo systemctl restart httpd
```

### With PHP Support

```bash
# Install PHP
sudo yum install -y httpd php php-mysqlnd php-fpm

# Create test PHP file
echo "<?php phpinfo(); ?>" | sudo tee /var/www/html/info.php

# Restart Apache
sudo systemctl restart httpd
```

### With Custom Configuration

Upload custom Apache configuration:

```json
{
  "provisioners": [
    {
      "type": "file",
      "source": "httpd.conf",
      "destination": "/tmp/httpd.conf"
    },
    {
      "type": "shell",
      "inline": [
        "sudo mv /tmp/httpd.conf /etc/httpd/conf/httpd.conf",
        "sudo systemctl restart httpd"
      ]
    }
  ]
}
```

## Troubleshooting

### Build Fails at Provisioning

**Check SSH connectivity:**
```bash
# Verify SSH key is correct
ssh -i ~/.ssh/id_rsa root@INSTANCE_IP

# Check instance status
ibmcloud pi instance INSTANCE_ID
```

### Apache Won't Start

**Check logs in the instance:**
```bash
sudo journalctl -u httpd -n 50
sudo tail -f /var/log/httpd/error_log
```

### Image Capture Fails

**Verify COS credentials:**
```bash
# Test with AWS CLI
aws configure set aws_access_key_id YOUR_ACCESS_KEY
aws configure set aws_secret_access_key YOUR_SECRET_KEY
aws s3 ls s3://your-bucket --endpoint-url=https://s3.us-south.cloud-object-storage.appdomain.cloud
```

### Timeout Issues

**Increase timeouts in template:**
```json
{
  "ssh_timeout": "30m",
  "cleanup_timeout": "15m"
}
```

## Cost Considerations

- **Instance Runtime**: Charged per hour while building
- **Storage**: COS storage for images
- **Network**: Data transfer costs
- **Cleanup**: Ensure resources are deleted after build

**Tip**: Use `cleanup_timeout` to ensure instances are deleted even if build fails.

## Best Practices

1. **Use Variables**: Keep sensitive data in variables
2. **Version Images**: Include timestamps in image names
3. **Test Scripts**: Test provisioning scripts separately first
4. **Monitor Builds**: Watch build progress for issues
5. **Clean Up**: Verify resources are deleted after builds
6. **Document Changes**: Keep track of image versions and changes

## Next Steps

- Customize the provisioning script for your needs
- Add more provisioners (Ansible, Chef, etc.)
- Implement multi-stage builds
- Integrate with CI/CD pipelines
- Create additional application-specific images

## Additional Resources

- [Main Documentation](../../../docs/README.md)
- [Troubleshooting Guide](../../../docs/TROUBLESHOOTING.md)
- [More Examples](../../)
- [Packer Documentation](https://www.packer.io/docs)

## Support

- **Issues**: [GitHub Issues](https://github.com/ppc64le-cloud/packer-plugin-powervs/issues)
- **Questions**: [GitHub Discussions](https://github.com/ppc64le-cloud/packer-plugin-powervs/discussions)

---

**Note**: Remember to replace placeholder values with your actual configuration before building.
