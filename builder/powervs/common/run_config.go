//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Source,COS,StockImage,Capture,CaptureCOS

package common

import (
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Source struct {
	Name       string      `mapstructure:"name" required:"false"`
	COS        *COS        `mapstructure:"cos" required:"false"`
	StockImage *StockImage `mapstructure:"stock_image" required:"false"`
}

type COS struct {
	Bucket string `mapstructure:"bucket" required:"true"`
	Object string `mapstructure:"object" required:"true"`
	Region string `mapstructure:"region" required:"true"`
}

type StockImage struct {
	Name string `mapstructure:"name" required:"true"`
}

type Capture struct {
	Name string      `mapstructure:"name" required:"true"`
	COS  *CaptureCOS `mapstructure:"cos" required:"false"`
}

type CaptureCOS struct {
	Bucket    string `mapstructure:"bucket" required:"true"`
	Region    string `mapstructure:"region" required:"true"`
	AccessKey string `mapstructure:"access_key" required:"true"`
	SecretKey string `mapstructure:"secret_key" required:"true"`
}

type RunConfig struct {
	InstanceName string  `mapstructure:"instance_name" required:"true"`
	KeyPairName  string  `mapstructure:"key_pair_name" required:"true"`
	Source       Source  `mapstructure:"source" required:"true"`
	Capture      Capture `mapstructure:"capture" required:"true"`

	// Communicator settings
	Comm communicator.Config `mapstructure:",squash"`
}

func (c *RunConfig) Prepare(ctx *interpolate.Context) []error {
	// Validation
	errs := c.Comm.Prepare(ctx)
	return errs
}
