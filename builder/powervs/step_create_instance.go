package powervs

import (
	"context"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

const (
	// DefaultCleanupTimeout is the default maximum time to wait for instance deletion
	DefaultCleanupTimeout = 10 * time.Minute
	// CleanupPollInterval is how often to check instance deletion status
	CleanupPollInterval = 10 * time.Second
)

type StepCreateInstance struct {
	InstanceName   string
	KeyPairName    string
	UserData       string
	CleanupTimeout time.Duration

	doCleanup bool
}

func (s *StepCreateInstance) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Creating Instance")

	instanceClient := state.Get("instanceClient").(*instance.IBMPIInstanceClient)

	net := state.Get("network").(*models.Network)

	imageRef := state.Get("source_image").(*models.ImageReference)

	networks := []*models.PVMInstanceAddNetwork{}

	if state.Get("networks") != nil {
		// Several subnets have been specified -> pass them all for vm creation
		networks = []*models.PVMInstanceAddNetwork{}

		for _, subnet := range state.Get("networks").([]string) {
			subnetAdd := &models.PVMInstanceAddNetwork{
				NetworkID: &subnet,
			}
			networks = append(networks, subnetAdd)
		}
	} else {
		networks = append(networks, &models.PVMInstanceAddNetwork{NetworkID: net.NetworkID})
	}

	body := &models.PVMInstanceCreate{
		ImageID:     imageRef.ImageID,
		KeyPairName: s.KeyPairName,
		Memory:      core.Float64Ptr(4),
		Networks:    networks,
		ProcType:    core.StringPtr("shared"),
		Processors:  core.Float64Ptr(0.5),
		ServerName:  &s.InstanceName,
		StorageType: *imageRef.StorageType,
		UserData:    b64.StdEncoding.EncodeToString([]byte(s.UserData)),
	}
	ui.Message("Creating Instance")
	ins, err := instanceClient.Create(body)
	if err != nil {
		ui.Error(fmt.Sprintf("failed to create instance: %v", err))
		state.Put("error", fmt.Errorf("failed to create instance: %w", err))
		return multistep.ActionHalt
	}

	var insIDs []string
	for _, in := range *ins {
		insID := in.PvmInstanceID
		insIDs = append(insIDs, *insID)
	}

	if len(insIDs) == 0 {
		ui.Error("insIDs list is empty")
		state.Put("error", errors.New("insIDs list is empty"))
		return multistep.ActionHalt
	}

	var in *models.PVMInstance

	//nolint:staticcheck // SA1015 this disable staticcheck for the next line
	if err := pollUntil(time.Tick(30*time.Second), time.After(5*time.Minute), func() (bool, error) {
		in, err = instanceClient.Get(insIDs[0])
		if err != nil || in == nil {
			ui.Message("No response or error encountered while retrieving the instance. Retrying...")
			return false, nil
		}
		return true, nil
	}); err != nil {
		ui.Error(fmt.Sprintf("failed to get instance: %v", err))
		state.Put("error", fmt.Errorf("failed to create instance: %w", err))
		return multistep.ActionHalt
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

	// Initiate deletion
	err := instanceClient.Delete(*i.PvmInstanceID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error initiating instance deletion. Please delete the instance manually: %s (ID: %s)",
			*i.ServerName, *i.PvmInstanceID))
		ui.Error(fmt.Sprintf("Deletion error: %v", err))
		return
	}

	// Wait for deletion with timeout
	timeout := s.CleanupTimeout
	if timeout == 0 {
		timeout = DefaultCleanupTimeout
	}

	begin := time.Now()
	errorStateCount := 0
	maxErrorStateRetries := 5 // Give some retries even in ERROR state

	//nolint:staticcheck // SA1015 this disable staticcheck for the next line
	err = pollUntil(time.Tick(CleanupPollInterval), time.After(timeout), func() (bool, error) {
		in, err := instanceClient.Get(*i.PvmInstanceID)

		// Instance not found means it was successfully deleted
		if err != nil {
			ui.Message("Instance deleted successfully")
			return true, nil
		}

		// Instance still exists, check its state
		currentState := *in.Status
		elapsed := time.Since(begin).Round(time.Second)
		ui.Message(fmt.Sprintf("VM still exists, state: %s (elapsed: %s)", currentState, elapsed))

		// If instance is in ERROR state, warn but continue for a few retries
		// Sometimes instances in ERROR state can still be deleted
		if currentState == "ERROR" {
			errorStateCount++
			if errorStateCount == 1 {
				ui.Message("Warning: Instance entered ERROR state during deletion")
				ui.Message("Continuing to wait for deletion (may require manual cleanup)")
			}
			if errorStateCount >= maxErrorStateRetries {
				ui.Error(fmt.Sprintf(
					"Instance has been in ERROR state for %d checks. Deletion may have failed.",
					errorStateCount))
				ui.Error(fmt.Sprintf(
					"Please verify instance status manually: %s (ID: %s)",
					*i.ServerName, *i.PvmInstanceID))
				// Don't return error, let timeout handle it
			}
		}

		return false, nil
	})

	if err != nil {
		// Timeout occurred
		ui.Error(fmt.Sprintf(
			"Timed out waiting for instance deletion after %s", timeout))
		ui.Error(fmt.Sprintf(
			"Instance may need manual cleanup: %s (ID: %s)",
			*i.ServerName, *i.PvmInstanceID))

		// Try to get final state
		if finalInstance, getErr := instanceClient.Get(*i.PvmInstanceID); getErr == nil {
			ui.Error(fmt.Sprintf("Final instance state: %s", *finalInstance.Status))
		}
	}
}
