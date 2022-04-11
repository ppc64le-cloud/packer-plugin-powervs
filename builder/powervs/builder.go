//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package powervs

import (
	"context"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	powervscommon "github.com/ppc64le-cloud/packer-plugin-powervs/builder/powervs/common"
)

const BuilderId = "packer.builder.powervs"

type Config struct {
	common.PackerConfig        `mapstructure:",squash"`
	powervscommon.AccessConfig `mapstructure:",squash"`
	powervscommon.ImageConfig  `mapstructure:",squash"`
	powervscommon.RunConfig    `mapstructure:",squash"`
	MockOption                 string `mapstructure:"mock"`

	ctx interpolate.Context
}

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) (generatedVars []string, warnings []string, err error) {
	b.config.ctx.Funcs = powervscommon.TemplateFuncs
	err = config.Decode(&b.config, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &b.config.ctx,
	}, raws...)
	if err != nil {
		return nil, nil, err
	}
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, b.config.RunConfig.Prepare(&b.config.ctx)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packer.LogSecretFilter.Set(b.config.APIKey)
	packer.LogSecretFilter.Set(b.config.Capture.COS.AccessKey)
	packer.LogSecretFilter.Set(b.config.Capture.COS.SecretKey)

	// Return the placeholder for the generated data that will become available to provisioners and post-processors.
	// If the builder doesn't generate any data, just return an empty slice of string: []string{}
	buildGeneratedData := []string{"GeneratedMockData"}
	return buildGeneratedData, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	session, err := b.config.Session()
	if err != nil {
		return nil, err
	}

	imageClient, err := b.config.ImageClient(ctx, b.config.ServiceInstanceID)
	if err != nil {
		return nil, err
	}

	jobClient, err := b.config.JobClient(ctx, b.config.ServiceInstanceID)
	if err != nil {
		return nil, err
	}

	instanceClient, err := b.config.InstanceClient(ctx, b.config.ServiceInstanceID)
	if err != nil {
		return nil, err
	}

	networkClient, err := b.config.NetworkClient(ctx, b.config.ServiceInstanceID)
	if err != nil {
		return nil, err
	}

	var steps []multistep.Step

	steps = append(steps,
		&StepSayConfig{
			MockConfig: b.config.MockOption,
		},
		&StepImageBaseImage{
			Source: b.config.Source,
		},
		&StepCreateNetwork{},
		&StepCreateInstance{
			InstanceName: b.config.InstanceName,
			KeyPairName:  b.config.KeyPairName,
		},
		&communicator.StepConnect{
			Config:    &b.config.RunConfig.Comm,
			Host:      powervscommon.SSHHost(),
			SSHPort:   powervscommon.Port(),
			SSHConfig: b.config.RunConfig.Comm.SSHConfigFunc(),
		},
		new(commonsteps.StepProvision),
		&StepPrepare{},
		&StepCaptureInstance{
			Capture: b.config.RunConfig.Capture,
		},
	)

	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("powervsSession", session)
	state.Put("imageClient", imageClient)
	state.Put("jobClient", jobClient)
	state.Put("instanceClient", instanceClient)
	state.Put("networkClient", networkClient)

	// Set the value of the generated data that will become available to provisioners.
	// To share the data with post-processors, use the StateData in the artifact.
	state.Put("generated_data", map[string]interface{}{
		"GeneratedMockData": "mock-build-data",
	})

	// Run!
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if err, ok := state.GetOk("error"); ok {
		return nil, err.(error)
	}

	artifact := &Artifact{
		// Add the builder generated data to the artifact StateData so that post-processors
		// can access them.
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
	}
	return artifact, nil
}
