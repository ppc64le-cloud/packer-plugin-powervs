# Contributing to Packer Plugin PowerVS

Thank you for your interest in contributing to the Packer Plugin for IBM Cloud Power Virtual Server! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Documentation](#documentation)
- [Submitting Changes](#submitting-changes)
- [Release Process](#release-process)
- [Getting Help](#getting-help)

## Code of Conduct

This project adheres to a code of conduct that all contributors are expected to follow. By participating, you are expected to uphold this code:

- **Be respectful**: Treat everyone with respect and consideration
- **Be collaborative**: Work together and help each other
- **Be inclusive**: Welcome newcomers and diverse perspectives
- **Be professional**: Focus on what is best for the community
- **Be constructive**: Provide helpful feedback and suggestions

## Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go**: Version 1.18 or higher ([installation guide](https://golang.org/doc/install))
- **Git**: For version control
- **Make**: For build automation
- **Packer**: Version 1.7.0 or higher ([installation guide](https://www.packer.io/downloads))
- **IBM Cloud Account**: For testing (optional but recommended)

### Finding Issues to Work On

Good places to start:

1. **Good First Issues**: Look for issues labeled `good first issue`
2. **Help Wanted**: Issues labeled `help wanted` need community support
3. **Documentation**: Improvements to docs are always welcome
4. **Bug Reports**: Help fix reported bugs

Browse our [issue tracker](https://github.com/ppc64le-cloud/packer-plugin-powervs/issues) to find something that interests you.

## Development Setup

### 1. Fork and Clone

Fork the repository on GitHub, then clone your fork:

```bash
git clone https://github.com/YOUR_USERNAME/packer-plugin-powervs.git
cd packer-plugin-powervs
```

### 2. Add Upstream Remote

Add the original repository as an upstream remote:

```bash
git remote add upstream https://github.com/ppc64le-cloud/packer-plugin-powervs.git
```

### 3. Install Dependencies

Install Go dependencies:

```bash
go mod download
```

### 4. Build and Install the Plugin

Build and install the plugin:

```bash
make install
```

This builds the plugin and installs it to your Packer plugins directory automatically.

**For Development/Testing:**

```bash
make dev
```

This is a quicker option that builds and copies the binary to `~/.packer.d/plugins/` for rapid testing during development.

**Build Only (without installing):**

```bash
make build
```

This creates a `packer-plugin-powervs` binary in the current directory without installing it.

### 6. Verify Installation

Create a test template and verify the plugin works:

```bash
cd example
packer init .
packer validate .
```

## Development Workflow

### 1. Create a Branch

Create a feature branch for your work:

```bash
git checkout -b feature/your-feature-name
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test improvements

### 2. Make Changes

Make your changes following our [coding standards](#coding-standards).

### 3. Test Your Changes

Run tests to ensure your changes work:

```bash
# Unit tests
go test ./...

# Specific package
go test ./builder/powervs/...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

### 4. Generate Code

If you modified struct tags or added new configuration fields, regenerate code:

```bash
go generate ./...
```

### 5. Format Code

Format your code before committing:

```bash
# Format all Go files
go fmt ./...

# Or use gofmt directly
gofmt -s -w .
```

### 6. Commit Changes

Write clear, descriptive commit messages:

```bash
git add .
git commit -m "Add feature: description of what you did"
```

**Commit Message Guidelines:**
- Use present tense ("Add feature" not "Added feature")
- Use imperative mood ("Move cursor to..." not "Moves cursor to...")
- Limit first line to 72 characters
- Reference issues and pull requests when relevant
- Provide detailed description in commit body if needed

Example:
```
Add support for custom instance profiles

- Implement instance profile configuration
- Add validation for profile parameters
- Update documentation with examples

Fixes #123
```

### 7. Push Changes

Push your branch to your fork:

```bash
git push origin feature/your-feature-name
```

### 8. Create Pull Request

Open a pull request on GitHub with:
- Clear title describing the change
- Detailed description of what and why
- Reference to related issues
- Screenshots/examples if applicable

## Coding Standards

### Go Style Guide

Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go.html).

### Key Principles

1. **Simplicity**: Write simple, readable code
2. **Consistency**: Follow existing patterns in the codebase
3. **Documentation**: Document exported functions and types
4. **Error Handling**: Always handle errors appropriately
5. **Testing**: Write tests for new functionality

### Code Organization

```go
// Package documentation
package powervs

import (
    // Standard library imports first
    "context"
    "fmt"
    
    // Third-party imports
    "github.com/hashicorp/packer-plugin-sdk/multistep"
    
    // Local imports
    powervscommon "github.com/ppc64le-cloud/packer-plugin-powervs/builder/powervs/common"
)

// Exported types and constants
type Builder struct {
    config Config
    runner multistep.Runner
}

// Exported functions
func (b *Builder) Run(ctx context.Context) error {
    // Implementation
}

// Unexported helper functions
func validateConfig(c *Config) error {
    // Implementation
}
```

### Documentation

Document all exported types, functions, and constants:

```go
// Builder implements the packer.Builder interface for PowerVS.
// It creates custom images on IBM Cloud Power Virtual Server by:
// 1. Creating a temporary instance
// 2. Provisioning the instance
// 3. Capturing the instance as an image
type Builder struct {
    config Config
    runner multistep.Runner
}

// Run executes the build process.
// It returns an Artifact containing the image details or an error if the build fails.
func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
    // Implementation
}
```

### Error Handling

Always handle errors appropriately:

```go
// Good
client, err := b.config.ImageClient(ctx, b.config.ServiceInstanceID)
if err != nil {
    return nil, fmt.Errorf("failed to create image client: %w", err)
}

// Bad - ignoring errors
client, _ := b.config.ImageClient(ctx, b.config.ServiceInstanceID)
```

### Configuration Structs

Use struct tags for configuration:

```go
//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config

type Config struct {
    // The IBM Cloud API key for authentication.
    APIKey string `mapstructure:"api_key" required:"true"`
    
    // The PowerVS service instance ID.
    ServiceInstanceID string `mapstructure:"service_instance_id" required:"true"`
    
    // Enable debug logging. Default: false
    Debug bool `mapstructure:"debug" required:"false"`
}
```

## Testing

### Unit Tests

Write unit tests for all new functionality:

```go
func TestBuilder_Prepare(t *testing.T) {
    tests := []struct {
        name    string
        config  map[string]interface{}
        wantErr bool
    }{
        {
            name: "valid config",
            config: map[string]interface{}{
                "api_key": "test-key",
                "service_instance_id": "test-id",
                "zone": "lon04",
            },
            wantErr: false,
        },
        {
            name: "missing api_key",
            config: map[string]interface{}{
                "service_instance_id": "test-id",
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            b := &Builder{}
            _, _, err := b.Prepare(tt.config)
            if (err != nil) != tt.wantErr {
                t.Errorf("Prepare() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Acceptance Tests

Acceptance tests require IBM Cloud credentials:

```bash
# Set environment variables
export IBM_API_KEY="your-api-key"
export POWERVS_SERVICE_INSTANCE_ID="your-instance-id"
export POWERVS_ZONE="lon04"

# Run acceptance tests
PACKER_ACC=1 go test -count 1 -v ./... -timeout=120m
```

**Note**: Acceptance tests create real resources and may incur costs.

### Test Coverage

Aim for good test coverage:

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out
```

### Testing Best Practices

1. **Test edge cases**: Test boundary conditions and error cases
2. **Use table-driven tests**: For testing multiple scenarios
3. **Mock external dependencies**: Use interfaces for testability
4. **Clean up resources**: Ensure tests clean up after themselves
5. **Parallel tests**: Use `t.Parallel()` when possible

## Documentation

### Code Documentation

- Document all exported types, functions, and constants
- Use complete sentences with proper punctuation
- Include examples for complex functionality
- Keep documentation up-to-date with code changes

### User Documentation

When adding features, update:

1. **README.md**: Main project documentation
2. **docs/README.md**: Detailed user guide
3. **Example configurations**: In `example/` and `builder/examples/`
4. **MDX files**: Component-specific documentation in `docs/`

### Documentation Style

- Use clear, concise language
- Provide examples for complex concepts
- Include code snippets with syntax highlighting
- Add diagrams for architectural concepts
- Keep formatting consistent

## Submitting Changes

### Pull Request Process

1. **Update documentation**: Ensure docs reflect your changes
2. **Add tests**: Include tests for new functionality
3. **Run tests**: Verify all tests pass
4. **Update CHANGELOG**: Add entry describing your changes
5. **Create PR**: Open pull request with clear description

### Pull Request Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Acceptance tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Documentation updated
- [ ] Tests pass locally
- [ ] Commit messages are clear
- [ ] CHANGELOG updated

## Related Issues
Fixes #(issue number)
```

### Review Process

1. **Automated checks**: CI/CD runs tests and linters
2. **Maintainer review**: At least one maintainer reviews the PR
3. **Address feedback**: Make requested changes
4. **Approval**: PR is approved by maintainer
5. **Merge**: Maintainer merges the PR

### After Your PR is Merged

1. **Delete your branch**: Clean up your feature branch
2. **Update your fork**: Sync with upstream
3. **Celebrate**: You've contributed to the project! 🎉

## Release Process

Releases are managed by maintainers:

1. **Version bump**: Update version in `version/version.go`
2. **Update CHANGELOG**: Document all changes
3. **Create tag**: Tag the release commit
4. **GitHub release**: Create release on GitHub
5. **Build artifacts**: GoReleaser builds binaries
6. **Publish**: Release is published

### Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backwards compatible)
- **PATCH**: Bug fixes (backwards compatible)

## Getting Help

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and general discussion
- **Slack**: [Kubernetes Slack #powervs](https://kubernetes.slack.com/messages/powervs)

### Asking Questions

When asking for help:

1. **Search first**: Check existing issues and documentation
2. **Be specific**: Provide details about your problem
3. **Include context**: Share relevant code, configs, and error messages
4. **Be patient**: Maintainers are volunteers

### Reporting Bugs

Include in bug reports:

- **Description**: Clear description of the bug
- **Steps to reproduce**: Detailed steps to reproduce the issue
- **Expected behavior**: What you expected to happen
- **Actual behavior**: What actually happened
- **Environment**: Packer version, plugin version, OS, etc.
- **Logs**: Relevant log output (use `PACKER_LOG=1`)
- **Configuration**: Sanitized template configuration

### Suggesting Features

Include in feature requests:

- **Use case**: Why you need this feature
- **Proposed solution**: How you think it should work
- **Alternatives**: Other solutions you've considered
- **Additional context**: Any other relevant information

## Recognition

Contributors are recognized in:

- **CHANGELOG**: For significant contributions
- **README**: In the acknowledgments section
- **GitHub**: Contributor badge on your profile

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.

## Thank You!

Your contributions make this project better for everyone. Thank you for taking the time to contribute!

---

**Questions?** Open an issue or reach out on Slack.

**Found a typo in this guide?** Submit a PR to fix it!