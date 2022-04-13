package powervs

import (
	"context"
	"fmt"
	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"time"
)

type StepCreateInstance struct {
	InstanceName string
	KeyPairName  string

	doCleanup bool
}

func (s *StepCreateInstance) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Creating Instance")

	instanceClient := state.Get("instanceClient").(*instance.IBMPIInstanceClient)

	net := state.Get("network").(*models.Network)

	imageRef := state.Get("source_image").(*models.ImageReference)
	networks := []*models.PVMInstanceAddNetwork{
		{
			NetworkID: net.NetworkID,
		},
	}
	body := &models.PVMInstanceCreate{
		ImageID:     imageRef.ImageID,
		KeyPairName: s.KeyPairName,
		Memory:      core.Float64Ptr(2),
		Networks:    networks,
		ProcType:    core.StringPtr("shared"),
		Processors:  core.Float64Ptr(0.5),
		ServerName:  &s.InstanceName,
		StorageType: *imageRef.StorageType,
	}
	ui.Message("Creating Instance")
	ins, err := instanceClient.Create(body)
	if err != nil {
		ui.Error(fmt.Sprintf("failed to create instance: %v", err))
		return multistep.ActionHalt
	}

	var insIDs []string
	for _, in := range *ins {
		insID := in.PvmInstanceID
		insIDs = append(insIDs, *insID)
	}

	if len(insIDs) == 0 {
		ui.Error("insIDs list is empty")
		return multistep.ActionHalt
	}

	in, err := instanceClient.Get(insIDs[0])
	if err != nil {
		ui.Error(fmt.Sprintf("failed to get instance: %v", err))
	}
	ui.Message(fmt.Sprintf("Instance Created, Name: %s, ID: %s", *in.ServerName, *in.PvmInstanceID))

	state.Put("instance", in)
	s.doCleanup = true

	return multistep.ActionContinue
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepCreateInstance) Cleanup(state multistep.StateBag) {
	if !s.doCleanup {
		return
	}
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Deleting the Instance")
	instanceClient := state.Get("instanceClient").(*instance.IBMPIInstanceClient)
	i := state.Get("instance").(*models.PVMInstance)
	err := instanceClient.Delete(*i.PvmInstanceID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up instance. Please delete the instance manually: %s", *i.ServerName))
	}
	for {
		in, err := instanceClient.Get(*i.PvmInstanceID)
		if err == nil {
			ui.Message(fmt.Sprintf("VM still exists, state: %s", *in.Status))
			time.Sleep(10 * time.Second)
			continue
		} else {
			ui.Message("instance deleted successfully")
			break
		}
	}
}
