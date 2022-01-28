//go:generate packer-sdc struct-markdown

package common

import (
	"context"
	"github.com/IBM-Cloud/power-go-client/clients/instance"
	ps "github.com/IBM-Cloud/power-go-client/ibmpisession"
	"github.com/IBM/go-sdk-core/v5/core"
)

type AccessConfig struct {
	// The api key used to communicate with IBM Cloud.
	APIKey string `mapstructure:"api_key" required:"true"`

	// Region of a Power VS.
	Region string `mapstructure:"region" required:"true"`

	// Zone of a Power VS.
	Zone string `mapstructure:"zone" required:"true"`

	// Account ID of a IBM Cloud account.
	AccountID string `mapstructure:"account_id" required:"true"`

	// Enable debug logging, Default `false`.
	Debug bool `mapstructure:"debug" required:"false"`

	// Power VS ServiceInstanceID
	ServiceInstanceID string `mapstructure:"service_instance_id" required:"true"`

	session *ps.IBMPISession
}

func (c *AccessConfig) Session() (*ps.IBMPISession, error) {
	if c.session != nil {
		return c.session, nil
	}

	authenticator := &core.IamAuthenticator{
		ApiKey: c.APIKey,
	}
	options := &ps.IBMPIOptions{
		Authenticator: authenticator,
		UserAccount:   c.AccountID,
		Region:        c.Region,
		Zone:          c.Zone,
		Debug:         c.Debug,
	}
	session, err := ps.NewIBMPISession(options)
	if err != nil {
		return nil, err
	}
	c.session = session
	return session, nil
}

func (c *AccessConfig) ImageClient(ctx context.Context, id string) (*instance.IBMPIImageClient, error) {
	session, err := c.Session()
	if err != nil {
		return nil, err
	}
	return instance.NewIBMPIImageClient(ctx, session, id), nil
}

func (c *AccessConfig) InstanceClient(ctx context.Context, id string) (*instance.IBMPIInstanceClient, error) {
	session, err := c.Session()
	if err != nil {
		return nil, err
	}
	return instance.NewIBMPIInstanceClient(ctx, session, id), nil
}

func (c *AccessConfig) NetworkClient(ctx context.Context, id string) (*instance.IBMPINetworkClient, error) {
	session, err := c.Session()
	if err != nil {
		return nil, err
	}
	return instance.NewIBMPINetworkClient(ctx, session, id), nil
}

func (c *AccessConfig) JobClient(ctx context.Context, id string) (*instance.IBMPIJobClient, error) {
	session, err := c.Session()
	if err != nil {
		return nil, err
	}
	return instance.NewIBMPIJobClient(ctx, session, id), nil
}
