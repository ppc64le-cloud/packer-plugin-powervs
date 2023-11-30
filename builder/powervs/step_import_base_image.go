package powervs

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/IBM-Cloud/power-go-client/clients/instance"
	"github.com/IBM-Cloud/power-go-client/power/models"
	"github.com/IBM/go-sdk-core/v5/core"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"

	"github.com/ppc64le-cloud/packer-plugin-powervs/builder/powervs/common"
)

var (
	// ImageStateACTIVE is the string representing an image in an active state.
	ImageStateACTIVE = "active"

	// ImageStateFailed is the string representing an image in a failed state.
	ImageStateFailed = "failed"
)

const (
	JobWaitThreshold        = 30 * time.Minute
	JobPollInterval         = 2 * time.Minute
	ImageImportThreshold    = 30 * time.Minute
	ImageImportPollInterval = 2 * time.Minute
	StorageTypeTier1        = "tier1"
)

var (
	BucketAccessPublic = "public"
)

type StepImageBaseImage struct {
	Source  common.Source
	cleanup bool
}

func (s *StepImageBaseImage) SetCleanup() {
	s.cleanup = true
}

func (s *StepImageBaseImage) GetCleanup() bool {
	return s.cleanup
}

func (s *StepImageBaseImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	ui.Say("Importing the Base Image")
	imageClient := state.Get("imageClient").(*instance.IBMPIImageClient)
	jobClient := state.Get("jobClient").(*instance.IBMPIJobClient)
	switch {
	case s.Source.COS != nil:
		ui.Message(fmt.Sprintf("Importing from the COS bucket: %+v\n", s.Source.COS))
		if s.Source.Name == "" {
			s1 := rand.NewSource(time.Now().UnixNano())
			s.Source.Name = fmt.Sprintf("%s-image-%d", s.Source.COS.Bucket, rand.New(s1).Intn(100))
		}
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
		s.SetCleanup()
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
					ui.Error("timed out while waiting for image to be imported")
					return multistep.ActionHalt
				}
				ui.Message(fmt.Sprintf("Sleeping for %s Minutes", JobPollInterval))
				time.Sleep(JobPollInterval)
			}
		}
	case s.Source.StockImage != nil:
		ui.Message(fmt.Sprintf("Importing from the Stock Image: %+v\n", s.Source.StockImage))
		stockImages, err := imageClient.GetAllStockImages(true, true)
		if err != nil {
			ui.Error(fmt.Sprintf("failed to GetAllStockImages: %+v", err))
			return multistep.ActionHalt
		}
		stockImageID := ""
		for _, si := range stockImages.Images {
			if *si.Name == s.Source.StockImage.Name {
				stockImageID = *si.ImageID
			}
		}
		if stockImageID == "" {
			ui.Error(fmt.Sprintf("failed to find a %s in StockImages: %+v", s.Source.StockImage.Name, stockImages.Images))
			return multistep.ActionHalt
		}
		ui.Message(fmt.Sprintf("Stock image found with id: %s\n", stockImageID))
		body := &models.CreateImage{
			ImageID: stockImageID,
			Source:  core.StringPtr("root-project"),
		}
		image, err := imageClient.Create(body)
		if err != nil {
			ui.Error(fmt.Sprintf("failed to import StockImage: %+v", err))
			return multistep.ActionHalt
		}
		s.SetCleanup()
		s.Source.Name = *image.Name
		begin := time.Now()
	loop2:
		for {
			img, err := imageClient.Get(*image.ImageID)
			if err != nil {
				ui.Error(fmt.Sprintf("failed to Get an image: %+v", err))
				return multistep.ActionHalt
			}
			ui.Message(fmt.Sprintf("Image state: %s", img.State))
			switch img.State {
			case ImageStateFailed:
				return multistep.ActionHalt
			case ImageStateACTIVE:
				break loop2
			default:
				if time.Since(begin) >= ImageImportThreshold {
					ui.Error("timed out while waiting for image to be imported")
					return multistep.ActionHalt
				}
				ui.Message(fmt.Sprintf("Sleeping for %s Minutes", ImageImportPollInterval))
				time.Sleep(ImageImportPollInterval)
			}
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

	if imageRef != nil {
		if imageRef.ImageID != nil {
			ui.Message(fmt.Sprintf("Image found with ID: %s", *imageRef.ImageID))
		}
		state.Put("source_image", imageRef)
		return multistep.ActionContinue
	} else {
		return multistep.ActionHalt
	}
}

// Cleanup can be used to clean up any artifact created by the step.
// A step's clean up always run at the end of a build, regardless of whether provisioning succeeds or fails.
func (s *StepImageBaseImage) Cleanup(state multistep.StateBag) {
	if !s.GetCleanup() {
		return
	}
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Deleting the Image")
	imageClient := state.Get("imageClient").(*instance.IBMPIImageClient)
	si := state.Get("source_image").(*models.ImageReference)
	err := imageClient.Delete(*si.ImageID)
	if err != nil {
		ui.Error(fmt.Sprintf(
			"Error cleaning up an image. Please delete an image manually: %s", *si.Name))
	}
	for {
		img, err := imageClient.Get(*si.ImageID)
		if err == nil {
			ui.Message(fmt.Sprintf("Image still exists, state: %s", img.State))
			time.Sleep(10 * time.Second)
			continue
		} else {
			ui.Message("image deleted successfully")
			break
		}
	}
}
