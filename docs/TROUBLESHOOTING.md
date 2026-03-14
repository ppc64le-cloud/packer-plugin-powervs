# Troubleshooting Guide

This guide helps you diagnose and resolve common issues when using the Packer Plugin for IBM Cloud Power Virtual Server.

## Table of Contents

- [General Troubleshooting](#general-troubleshooting)
- [Authentication Issues](#authentication-issues)
- [Network Issues](#network-issues)
- [Image Issues](#image-issues)
- [Instance Issues](#instance-issues)
- [Build Failures](#build-failures)
- [Performance Issues](#performance-issues)
- [Debug Mode](#debug-mode)
- [Common Error Messages](#common-error-messages)
- [Getting Help](#getting-help)

## General Troubleshooting

### Enable Debug Logging

Enable detailed logging to diagnose issues:

```bash
# Set environment variable
export PACKER_LOG=1
export PACKER_LOG_PATH=packer.log

# Run your build
packer build template.pkr.hcl
```

In your template, enable plugin debug mode:

```hcl
source "powervs" "example" {
  debug = true
  # ... other configuration
}
```

### Verify Plugin Installation

Check if the plugin is correctly installed:

```bash
# List installed plugins
ls -la ~/.packer.d/plugins/

# Verify plugin version
packer plugins installed
```

### Validate Template

Always validate your template before building:

```bash
packer validate template.pkr.hcl
```

### Check Packer Version

Ensure you're using a compatible Packer version:

```bash
packer version
# Should be >= 1.7.0
```

## Authentication Issues

### Error: "Invalid API Key"

**Symptoms:**
```
Error: authentication failed: invalid API key
```

**Solutions:**

1. **Verify API Key Format:**
   ```bash
   # API key should be 44 characters
   echo $IBM_API_KEY | wc -c
   ```

2. **Check API Key Permissions:**
   - Ensure the API key has PowerVS access
   - Verify the key hasn't expired
   - Check if the key is for the correct account

3. **Test API Key:**
   ```bash
   # Install IBM Cloud CLI
   ibmcloud login --apikey $IBM_API_KEY
   
   # List PowerVS instances
   ibmcloud pi service-list
   ```

4. **Regenerate API Key:**
   - Go to IBM Cloud Console → Manage → Access (IAM) → API keys
   - Create a new API key with PowerVS permissions

### Error: "Service Instance Not Found"

**Symptoms:**
```
Error: service instance 'xxx' not found
```

**Solutions:**

1. **Verify Service Instance ID:**
   ```bash
   # List PowerVS instances
   ibmcloud pi service-list
   
   # Get instance details
   ibmcloud pi service-target <instance-name>
   ```

2. **Check Zone Configuration:**
   ```hcl
   # Ensure zone matches your service instance
   zone = "lon04"  # Must match instance location
   ```

3. **Verify Account Access:**
   - Ensure your API key has access to the account
   - Check if you're using the correct account ID

### Error: "Unauthorized Access"

**Symptoms:**
```
Error: 403 Forbidden - unauthorized access
```

**Solutions:**

1. **Check IAM Permissions:**
   Required permissions:
   - PowerVS Service: Editor or Administrator
   - Resource Group: Viewer (minimum)

2. **Verify Service ID:**
   If using a service ID, ensure it has the correct access policies

3. **Check Resource Group:**
   ```hcl
   # Specify resource group if needed
   account_id = "your-account-id"
   ```

## Network Issues

### Error: "SSH Connection Timeout"

**Symptoms:**
```
Error: timeout waiting for SSH connection
```

**Solutions:**

1. **Verify SSH Key:**
   ```bash
   # Check if key exists
   ls -la ~/.ssh/id_rsa
   
   # Verify key is uploaded to PowerVS
   ibmcloud pi key-list
   ```

2. **Check Network Configuration:**
   ```hcl
   # Use DHCP for automatic network setup
   dhcp_network = true
   
   # Or verify subnet IDs
   subnet_ids = ["valid-subnet-id"]
   ```

3. **Increase SSH Timeout:**
   ```hcl
   ssh_timeout = "30m"  # Increase from default
   ```

4. **Verify Security Groups:**
   - Ensure port 22 is open
   - Check firewall rules

5. **Test SSH Manually:**
   ```bash
   # Get instance IP from Packer logs
   ssh -i ~/.ssh/id_rsa root@<instance-ip>
   ```

### Error: "Network Not Found"

**Symptoms:**
```
Error: network with ID 'xxx' not found
```

**Solutions:**

1. **List Available Networks:**
   ```bash
   ibmcloud pi networks --instance-id <service-instance-id>
   ```

2. **Use DHCP Instead:**
   ```hcl
   dhcp_network = true
   # Remove subnet_ids
   ```

3. **Verify Subnet IDs:**
   ```hcl
   # Ensure subnet IDs are correct
   subnet_ids = ["correct-subnet-id-1", "correct-subnet-id-2"]
   ```

### Error: "DHCP Network Creation Failed"

**Symptoms:**
```
Error: failed to create DHCP network
```

**Solutions:**

1. **Check DHCP Service:**
   - Verify DHCP service is available in your zone
   - Some zones may not support DHCP

2. **Use Existing Network:**
   ```hcl
   dhcp_network = false
   subnet_ids = ["existing-subnet-id"]
   ```

3. **Check Quota:**
   - Verify you haven't exceeded network quota
   - Check PowerVS service limits

## Image Issues

### Error: "Image Import Failed"

**Symptoms:**
```
Error: failed to import image from Cloud Object Storage
```

**Solutions:**

1. **Verify COS Configuration:**
   ```hcl
   source {
     name = "my-image"
     cos {
       bucket = "correct-bucket-name"
       object = "image.ova.gz"  # Must exist
       region = "us-south"      # Must match bucket region
     }
   }
   ```

2. **Check Image Format:**
   - Supported formats: OVA, OVA.GZ
   - Verify file is not corrupted
   - Check file size (must be reasonable)

3. **Verify COS Access:**
   ```bash
   # Install AWS CLI
   aws configure set aws_access_key_id <access-key>
   aws configure set aws_secret_access_key <secret-key>
   
   # List bucket contents
   aws s3 ls s3://bucket-name --endpoint-url=https://s3.us-south.cloud-object-storage.appdomain.cloud
   ```

4. **Check Permissions:**
   - Ensure COS bucket is accessible
   - Verify HMAC credentials are correct
   - Check bucket policies

### Error: "Stock Image Not Found"

**Symptoms:**
```
Error: stock image 'xxx' not found
```

**Solutions:**

1. **List Available Images:**
   ```bash
   ibmcloud pi images --instance-id <service-instance-id>
   ```

2. **Use Correct Image Name:**
   ```hcl
   source {
     stock_image {
       name = "CentOS-Stream-8"  # Exact name from list
     }
   }
   ```

3. **Check Zone Availability:**
   - Not all images are available in all zones
   - Try a different zone or image

### Error: "Image Capture Failed"

**Symptoms:**
```
Error: failed to capture instance as image
```

**Solutions:**

1. **Check Capture Configuration:**
   ```hcl
   capture {
     name = "unique-image-name"
     destination = "cloud-storage"  # or "image-catalog" or "both"
     cos {
       bucket     = "valid-bucket"
       region     = "us-south"
       access_key = var.access_key
       secret_key = var.secret_key
     }
   }
   ```

2. **Verify COS Credentials:**
   - Test HMAC credentials separately
   - Ensure bucket exists and is accessible
   - Check bucket region matches configuration

3. **Check Disk Space:**
   - Ensure sufficient space in COS bucket
   - Verify PowerVS storage quota

4. **Wait for Instance to Stabilize:**
   - Add a pause before capture:
   ```hcl
   provisioner "shell" {
     inline = ["sleep 60"]  # Wait before capture
   }
   ```

## Instance Issues

### Error: "Instance Creation Failed"

**Symptoms:**
```
Error: failed to create instance
```

**Solutions:**

1. **Check Instance Configuration:**
   ```hcl
   instance_name = "unique-name-${timestamp()}"
   key_pair_name = "valid-ssh-key"
   ```

2. **Verify Resource Quota:**
   - Check if you've exceeded instance quota
   - Verify available resources in zone

3. **Check Instance Name:**
   - Must be unique
   - Use timestamp for uniqueness:
   ```hcl
   instance_name = "packer-${timestamp()}"
   ```

4. **Verify SSH Key:**
   ```bash
   # List available keys
   ibmcloud pi keys --instance-id <service-instance-id>
   ```

### Error: "Instance Deletion Timeout"

**Symptoms:**
```
Error: timeout waiting for instance deletion
```

**Solutions:**

1. **Increase Cleanup Timeout:**
   ```hcl
   cleanup_timeout = "20m"  # Increase from default 10m
   ```

2. **Manual Cleanup:**
   ```bash
   # List instances
   ibmcloud pi instances --instance-id <service-instance-id>
   
   # Delete stuck instance
   ibmcloud pi instance-delete <instance-id>
   ```

3. **Check Instance State:**
   - Instance may be in a transitional state
   - Wait and retry

### Error: "Instance Not Ready"

**Symptoms:**
```
Error: instance not in active state
```

**Solutions:**

1. **Increase Wait Time:**
   ```hcl
   # Add longer timeout
   ssh_timeout = "30m"
   ```

2. **Check Instance Logs:**
   ```bash
   # View instance console
   ibmcloud pi instance-console <instance-id>
   ```

3. **Verify Image Compatibility:**
   - Ensure base image is compatible
   - Check if image supports cloud-init

## Build Failures

### Error: "Provisioner Failed"

**Symptoms:**
```
Error: provisioner "shell" failed
```

**Solutions:**

1. **Check Provisioner Script:**
   ```hcl
   provisioner "shell" {
     inline = [
       "set -e",  # Exit on error
       "yum update -y",
       "yum install -y httpd"
     ]
   }
   ```

2. **Add Error Handling:**
   ```hcl
   provisioner "shell" {
     inline = ["command || true"]  # Continue on error
   }
   ```

3. **Use on_error:**
   ```hcl
   provisioner "shell" {
     inline = ["risky-command"]
     on_error = "continue"  # or "abort" or "ask"
   }
   ```

4. **Debug Provisioner:**
   ```hcl
   provisioner "shell" {
     inline = [
       "set -x",  # Enable debug output
       "your-commands"
     ]
   }
   ```

### Error: "Build Timeout"

**Symptoms:**
```
Error: build exceeded timeout
```

**Solutions:**

1. **Increase Timeout:**
   ```bash
   packer build -timeout=2h template.pkr.hcl
   ```

2. **Optimize Provisioning:**
   - Combine commands to reduce SSH overhead
   - Use package caching
   - Parallelize when possible

3. **Check Network Speed:**
   - Slow downloads can cause timeouts
   - Use local mirrors when possible

## Performance Issues

### Slow Builds

**Solutions:**

1. **Use DHCP Networks:**
   ```hcl
   dhcp_network = true  # Faster than custom networks
   ```

2. **Optimize Provisioning:**
   ```hcl
   provisioner "shell" {
     inline = [
       # Combine commands
       "yum update -y && yum install -y package1 package2 package3"
     ]
   }
   ```

3. **Use Package Caching:**
   ```hcl
   provisioner "shell" {
     inline = [
       "yum install -y yum-plugin-fastestmirror",
       "yum makecache fast"
     ]
   }
   ```

4. **Parallel Builds:**
   ```hcl
   build {
     sources = [
       "source.powervs.image1",
       "source.powervs.image2"
     ]
     # Builds run in parallel
   }
   ```

### High Costs

**Solutions:**

1. **Minimize Build Time:**
   - Optimize provisioning scripts
   - Use efficient base images
   - Clean up resources promptly

2. **Use Image Catalog:**
   ```hcl
   capture {
     destination = "image-catalog"  # Avoid COS costs
   }
   ```

3. **Monitor Resources:**
   ```bash
   # Check running instances
   ibmcloud pi instances --instance-id <service-instance-id>
   ```

## Debug Mode

### Enable Comprehensive Debugging

```bash
# Set all debug flags
export PACKER_LOG=1
export PACKER_LOG_PATH=packer-debug.log
export TF_LOG=DEBUG
```

```hcl
source "powervs" "debug" {
  debug = true
  
  # Add verbose SSH logging
  ssh_timeout = "30m"
  
  # ... other configuration
}
```

### Analyze Debug Logs

```bash
# View logs in real-time
tail -f packer-debug.log

# Search for errors
grep -i error packer-debug.log

# Search for specific operations
grep -i "creating instance" packer-debug.log
```

## Common Error Messages

### "context deadline exceeded"

**Cause:** Operation timed out

**Solution:** Increase timeout values or check network connectivity

### "connection refused"

**Cause:** Service not accessible

**Solution:** Check network configuration and firewall rules

### "resource not found"

**Cause:** Referenced resource doesn't exist

**Solution:** Verify resource IDs and names

### "quota exceeded"

**Cause:** Account limits reached

**Solution:** Request quota increase or clean up resources

### "invalid parameter"

**Cause:** Configuration value is invalid

**Solution:** Check parameter format and allowed values

## Getting Help

### Before Asking for Help

1. **Check this guide**: Review relevant sections
2. **Search issues**: Look for similar problems
3. **Enable debug logging**: Gather detailed logs
4. **Simplify**: Create minimal reproduction case

### Where to Get Help

- **Report Bugs**: [GitHub Issues](https://github.com/ppc64le-cloud/packer-plugin-powervs/issues) - For bug reports and feature requests
- **Ask Questions**: [GitHub Discussions](https://github.com/ppc64le-cloud/packer-plugin-powervs/discussions) - For questions and community support
- **Documentation**: [docs/README.md](README.md) - For guides and references

### Information to Include When Reporting Issues

When opening a GitHub issue, include:

1. **Environment:**
   - Packer version: `packer version`
   - Plugin version
   - Operating system
   - Go version (if building from source)

2. **Configuration:**
   - Sanitized template (remove credentials)
   - Variable values (remove sensitive data)

3. **Logs:**
   - Full debug logs (`PACKER_LOG=1`)
   - Error messages
   - Stack traces

4. **Steps to Reproduce:**
   - Detailed steps to reproduce the issue
   - Expected vs actual behavior

5. **Context:**
   - What you're trying to accomplish
   - What you've already tried

### Example Issue Report

```markdown
## Description
Build fails when capturing image to Cloud Object Storage

## Environment
- Packer: 1.8.0
- Plugin: 0.0.1
- OS: macOS 12.0
- Zone: lon04

## Configuration
```hcl
source "powervs" "example" {
  api_key = var.api_key
  service_instance_id = "xxx"
  zone = "lon04"
  
  capture {
    destination = "cloud-storage"
    cos {
      bucket = "my-bucket"
      region = "us-south"
      access_key = var.access_key
      secret_key = var.secret_key
    }
  }
}
```

## Error Message
```
Error: failed to capture instance: invalid credentials
```

## Steps to Reproduce
1. Run `packer build template.pkr.hcl`
2. Build completes provisioning
3. Capture fails with credentials error

## What I've Tried
- Verified COS credentials work with AWS CLI
- Checked bucket exists and is accessible
- Tried different regions

## Debug Logs
[Attach packer-debug.log]
```

---

**Still stuck?** Don't hesitate to ask for help! The community is here to support you.