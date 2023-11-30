package powervs

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

const (
	PrepareWaitThreshold = 6 * time.Minute
	PreparePollInterval  = 2 * time.Minute
)

type StepPrepare struct {
}

func (s *StepPrepare) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Preparing Instance")
	instanceClient := state.Get("instanceClient").(*instance.IBMPIInstanceClient)
	i := state.Get("instance").(*models.PVMInstance)
	body := &models.PVMInstanceAction{
		Action: core.StringPtr("stop"),
	}
	err := instanceClient.Action(*i.PvmInstanceID, body)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error stopping the instance, %v", err))
		return multistep.ActionHalt
	}
	begin := time.Now()
	for {
		in, err := instanceClient.Get(*i.PvmInstanceID)
		if err != nil {
			ui.Error(fmt.Sprintf("failed to get instane, err: %+v", err))
			return multistep.ActionHalt
		}
		if *in.Status == "SHUTOFF" {
			return multistep.ActionContinue
		} else if time.Since(begin) >= PrepareWaitThreshold {
			ui.Error("timed out waiting for vm to shutoff")
			return multistep.ActionHalt
		}
		time.Sleep(PreparePollInterval)
	}
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepPrepare) Cleanup(_ multistep.StateBag) {
	// Nothing to clean
}
