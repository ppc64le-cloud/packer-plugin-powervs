package powervs

import (
	"context"
	"fmt"
	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type StepCreateNetwork struct {
	doCleanup bool
}

func (s *StepCreateNetwork) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Creating Instance")

	networkClient := state.Get("networkClient").(*instance.IBMPINetworkClient)

	netBody := &models.NetworkCreate{
		DNSServers: []string{"8.8.8.8", "9.9.9.9"},
		Type:       core.StringPtr("pub-vlan"),
	}
	ui.Message("Creating Network")
	net, err := networkClient.Create(netBody)
	if err != nil {
		ui.Error(fmt.Sprintf("failed to create network: %v", err))
		return multistep.ActionHalt
	}
	ui.Message(fmt.Sprintf("Network Created, Name: %s, ID: %s", *net.Name, *net.NetworkID))
	state.Put("network", net)
	s.doCleanup = true

	return multistep.ActionContinue
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepCreateNetwork) Cleanup(state multistep.StateBag) {
	if !s.doCleanup {
		return
	}
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Deleting the Network")
	networkClient := state.Get("networkClient").(*instance.IBMPINetworkClient)
	net := state.Get("network").(*models.Network)
	err := networkClient.Delete(*net.NetworkID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up network. Please delete the network manually: %s", *net.Name))
	}
}
