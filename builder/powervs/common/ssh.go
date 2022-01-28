package common

import (
	"errors"
	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"time"
)

var (
	// modified in tests
	sshHostSleepDuration = time.Minute
)

// SSHHost returns a function that can be given to the SSH communicator
// for determining the SSH address of the instance.
func SSHHost() func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		instanceClient := state.Get("instanceClient").(*instance.IBMPIInstanceClient)
		host := ""
		const tries = 15
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
