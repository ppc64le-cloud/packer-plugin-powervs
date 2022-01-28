package powervs

import (
	"context"
	"fmt"
	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/ppc64le-cloud/packer-plugin-powervs/builder/powervs/common"
	"time"
)

const (
	JobWaitThreshold = 30 * time.Minute
	JobPollInterval  = 2 * time.Minute
	StorageTypeTier1 = "tier1"
)

var (
	BucketAccessPublic = "public"
)

type StepImageBaseImage struct {
	Source common.Source
}

func (s *StepImageBaseImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Importing the Base Image")
	imageClient := state.Get("imageClient").(*instance.IBMPIImageClient)
	jobClient := state.Get("jobClient").(*instance.IBMPIJobClient)
	if s.Source.Import {
		switch {
		case s.Source.COS != nil:
			ui.Message(fmt.Sprintf("Importing from the COS bucket: %+v\n", s.Source.COS))
			body := &models.CreateCosImageImportJob{
				ImageName:     &s.Source.Name,
				BucketName:    core.StringPtr(s.Source.COS.Bucket),
				BucketAccess:  &BucketAccessPublic,
				Region:        core.StringPtr(s.Source.COS.Region),
				ImageFilename: core.StringPtr(s.Source.COS.Object),
				StorageType:   StorageTypeTier1,
			}
			imageJob, err := imageClient.CreateCosImage(body)
			if err != nil {
				ui.Error(fmt.Sprintf("failed to CreateCosImage: %+v", err))
				return multistep.ActionHalt
			}
			begin := time.Now()
		loop:
			for {
				job, err := jobClient.Get(*imageJob.ID)
				ui.Message(fmt.Sprintf("Job state: %s, progress: %s, message: %s", *job.Status.State, *job.Status.Progress, job.Status.Message))
				if err != nil {
					ui.Error(fmt.Sprintf("failed to Get Import Job: %+v", err))
					return multistep.ActionHalt
				}
				switch *job.Status.State {
				case "failed":
					return multistep.ActionHalt
				case "completed":
					break loop
				default:
					if time.Since(begin) >= JobWaitThreshold {
						ui.Error(fmt.Sprintf("timed out while waiting for image to be imported"))
						return multistep.ActionHalt
					}
					ui.Message(fmt.Sprintf("Sleeping for %s Minutes", JobPollInterval))
					time.Sleep(JobPollInterval)
				}
			}
		case s.Source.StockImage != nil:
			//TODO
		}
	}

	var imageRef *models.ImageReference
	images, err := imageClient.GetAll()
	if err != nil {
		ui.Error(fmt.Sprintf("failed to get all the images: %v", err))
		return multistep.ActionHalt
	}
	for _, image := range images.Images {
		if *image.Name == s.Source.Name {
			imageRef = image
		}
	}
	ui.Message(fmt.Sprintf("Image found with ID: %s", *imageRef.ImageID))

	if imageRef != nil {
		state.Put("source_image", imageRef)
		return multistep.ActionContinue
	} else {
		return multistep.ActionHalt
	}
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepImageBaseImage) Cleanup(_ multistep.StateBag) {
	// Nothing to clean
}
