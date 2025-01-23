package powervs

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/ppc64le-cloud/packer-plugin-powervs/builder/powervs/common"
)

const (
	CaptureJobWaitThreshold = 1 * time.Hour
	CaptureJobPollInterval  = 5 * time.Minute
)

var (
	CaptureDestinationDefault = "cloud-storage"
)

type StepCaptureInstance struct {
	Capture common.Capture
}

func (s *StepCaptureInstance) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Capturing Instance")

	instanceClient := state.Get("instanceClient").(*instance.IBMPIInstanceClient)
	i := state.Get("instance").(*models.PVMInstance)

	in, err := instanceClient.Get(*i.PvmInstanceID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"failed to get instance: %s, err: %v", *in.ServerName, err))
		return multistep.ActionHalt
	}

	captureDestination := CaptureDestinationDefault
	if s.Capture.Destination != "" {
		captureDestination = s.Capture.Destination
	}

	body := &models.PVMInstanceCapture{
		CaptureDestination: &captureDestination,
		CaptureName:        &s.Capture.Name,
	}
	if s.Capture.COS != nil {
		body.CloudStorageAccessKey = s.Capture.COS.AccessKey
		body.CloudStorageImagePath = s.Capture.COS.Bucket
		body.CloudStorageRegion = s.Capture.COS.Region
		body.CloudStorageSecretKey = s.Capture.COS.SecretKey
	}
	jobRef, err := instanceClient.CaptureInstanceToImageCatalogV2(*i.PvmInstanceID, body)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"failed to capture instance: %s, err: %v", *in.ServerName, err))
		return multistep.ActionHalt
	}

	jobClient := state.Get("jobClient").(*instance.IBMPIJobClient)
	begin := time.Now()
loop:
	for {
		job, err := jobClient.Get(*jobRef.ID)
		ui.Message(fmt.Sprintf("Job state: %s, progress: %s, message: %s", *job.Status.State, *job.Status.Progress, job.Status.Message))
		if err != nil {
			ui.Error(fmt.Sprintf("failed to Get capture Job: %+v", err))
			return multistep.ActionHalt
		}
		switch *job.Status.State {
		case "failed":
			return multistep.ActionHalt
		case "completed":
			break loop
		default:
			if time.Since(begin) >= CaptureJobWaitThreshold {
				ui.Error(fmt.Sprintf("timed out while waiting for image to be captured"))
				return multistep.ActionHalt
			}
			ui.Message(fmt.Sprintf("Sleeping for %s", CaptureJobPollInterval))
			time.Sleep(CaptureJobPollInterval)
		}
	}

	return multistep.ActionContinue
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepCaptureInstance) Cleanup(_ multistep.StateBag) {
	// Nothing to clean
}
