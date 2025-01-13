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
	// DHCPServerActiveTimeOut is time to wait for DHCP Server status to become active.
	DHCPServerActiveTimeOut = 15 * time.Minute
	// DHCPServerInterval is time to sleep before checking DHCP Server status.
	DHCPServerInterval = 1 * time.Minute
)

type StepCreateNetwork struct {
	SubnetID    string
	DHCPNetwork bool
	doCleanup   bool
}

func (s *StepCreateNetwork) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

	networkClient := state.Get("networkClient").(*instance.IBMPINetworkClient)

	if s.SubnetID != "" {
		ui.Say("The subnet is specified by the user; reuse it instead of creating a new one.")
		net, err := networkClient.Get(s.SubnetID)
		if err != nil {
			ui.Error(fmt.Sprintf("failed to get subnet: %s, error: %v", s.SubnetID, err))
			return multistep.ActionHalt
		}
		ui.Message(fmt.Sprintf("Network found!, Name: %s, ID: %s", *net.Name, *net.NetworkID))
		state.Put("network", net)
		// do not delete the user specified subnet, hence skipping the cleanup
		s.doCleanup = false
		return multistep.ActionContinue
	}

	// If CreateDHCPNetwork is set, Create DHCP network.
	if s.DHCPNetwork {
		ui.Say("Creating DHCP network")
		if err := s.createDHCPNetwork(state); err != nil {
			ui.Error(fmt.Sprintf("failed to create DHCP network: %v", err))
			return multistep.ActionHalt
		}
		s.doCleanup = true
		return multistep.ActionContinue
	}

	ui.Say("Creating network")
	netBody := &models.NetworkCreate{
		DNSServers: []string{"8.8.8.8", "9.9.9.9"},
		Type:       core.StringPtr("pub-vlan"),
	}
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

	if s.DHCPNetwork {
		ui.Message("Deleting DHCP server")
		dhcpServerID := state.Get("dhcpServerID").(string)
		dhcpClient := state.Get("dhcpClient").(*instance.IBMPIDhcpClient)

		if err := dhcpClient.Delete(dhcpServerID); err != nil {
			ui.Error(fmt.Sprintf("Error cleaning up DHCP server. Please delete the DHCP server manually: %s error: %v", dhcpServerID, err.Error()))
		}
		ui.Message("Successfully deleted DHCP server")
		return
	}
	networkClient := state.Get("networkClient").(*instance.IBMPINetworkClient)
	net := state.Get("network").(*models.Network)
	err := networkClient.Delete(*net.NetworkID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up network. Please delete the network manually: %s", *net.Name))
	}
	ui.Message("Successfully deleted network")
}

func (s *StepCreateNetwork) createDHCPNetwork(state multistep.StateBag) error {
	ui := state.Get("ui").(packersdk.Ui)
	dhcpClient := state.Get("dhcpClient").(*instance.IBMPIDhcpClient)

	dhcpServer, err := dhcpClient.Create(&models.DHCPServerCreate{})
	if err != nil {
		return fmt.Errorf("error failed to create DHCP server: %v", err)
	}

	if dhcpServer.ID == nil {
		return fmt.Errorf("error created DHCP server ID is nil")
	}
	state.Put("dhcpServerID", *dhcpServer.ID)

	startTime := time.Now()
	var networkID string
	for {
		dhcpServerDetails, err := dhcpClient.Get(*dhcpServer.ID)
		if err != nil {
			return err
		}
		if dhcpServerDetails.Network != nil && dhcpServerDetails.Network.ID != nil {
			networkID = *dhcpServerDetails.Network.ID
			ui.Message("DHCP server in active state")
			break
		}
		if time.Since(startTime) > DHCPServerActiveTimeOut {
			return fmt.Errorf("error DHCP server did not become active even after %f min", DHCPServerActiveTimeOut.Minutes())
		}
		ui.Message("Wating for DHCP server to become active")
		time.Sleep(DHCPServerInterval)
	}
	ui.Say("Fetching network details")
	networkClient := state.Get("networkClient").(*instance.IBMPINetworkClient)
	// fetch the dhcp network details and store it for future usage.
	net, err := networkClient.Get(networkID)
	if err != nil {
		return fmt.Errorf("error fetching network details with network id %s error: %v", networkID, err)
	}
	state.Put("network", net)
	return nil
}
