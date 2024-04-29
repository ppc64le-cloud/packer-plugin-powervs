package common

import (
	"errors"
	"fmt"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

var (
	// modified in tests
	sshHostSleepDuration = time.Minute
)

// SSHHost returns a function that can be given to the SSH communicator
// for determining the SSH address of the instance.
func SSHHost() func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		ui := state.Get("ui").(packersdk.Ui)
		ui.Message("Fetching IP for machine")
		instanceClient := state.Get("instanceClient").(*instance.IBMPIInstanceClient)
		host := ""
		const tries = 25
		for j := 0; j <= tries; j++ {
			i := state.Get("instance").(*models.PVMInstance)
			in, err := instanceClient.Get(*i.PvmInstanceID)
			if err != nil {
				return "", errors.New("couldn't determine address for instance: failed to get instance")
			}
			for _, net := range in.Networks {
				if net.ExternalIP != "" {
					host = net.ExternalIP
				}
			}
			if host != "" {
				return host, nil
			}

			dhcpServerID, ok := state.GetOk("dhcpServerID")
			if !ok {
				// if the dhcpServerID is not set, dont try to fetch IP from DHCP server, instead wait for address to get populated.
				ui.Message("Machine IP is not yet found, Trying again")
				time.Sleep(sshHostSleepDuration)
				continue
			}
			dhcpClient := state.Get("dhcpClient").(*instance.IBMPIDhcpClient)
			ui.Message("Getting Instance IP from DHCP server")

			net := state.Get("network").(*models.Network)
			networkID := net.NetworkID

			var pvmNetwork *models.PVMInstanceNetwork
			for _, network := range in.Networks {
				if network.NetworkID == *networkID {
					pvmNetwork = network
					ui.Message("Found network attached to VM")
				}
			}

			if pvmNetwork == nil {
				ui.Message("Failed to get network attached to VM, Trying again")
				time.Sleep(sshHostSleepDuration)
				continue
			}

			dhcpServerDetails, err := dhcpClient.Get(dhcpServerID.(string))
			if err != nil {
				ui.Error(fmt.Sprintf("Failed to get DHCP server details: %v", err))
				return "", err
			}

			if dhcpServerDetails == nil {
				ui.Error(fmt.Sprintf("DHCP server details is nil, DHCPServerID: %s", dhcpServerID))
				return "", err
			}

			var internalIP string
			for _, lease := range dhcpServerDetails.Leases {
				if *lease.InstanceMacAddress == pvmNetwork.MacAddress {
					ui.Message(fmt.Sprintf("Found internal ip for VM from DHCP lease IP %s", *lease.InstanceIP))
					internalIP = *lease.InstanceIP
					break
				}
			}
			if internalIP != "" {
				return internalIP, nil
			}

			ui.Message("Machine IP is not yet found from DHCP server lease, Trying again")
			time.Sleep(sshHostSleepDuration)
		}
		return "", errors.New("couldn't determine address for instance")
	}
}

// Port returns a function that can be given to the SSH communicator
// for determining the SSH Port
func Port() func(multistep.StateBag) (int, error) {
	return func(state multistep.StateBag) (int, error) {
		return 22, nil
	}
}
