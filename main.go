package main

import (
	"fmt"
	"github.com/ppc64le-cloud/packer-plugin-powervs/builder/powervs"
	powervsData "github.com/ppc64le-cloud/packer-plugin-powervs/datasource/powervs"
	powervsPP "github.com/ppc64le-cloud/packer-plugin-powervs/post-processor/powervs"
	powervsProv "github.com/ppc64le-cloud/packer-plugin-powervs/provisioner/powervs"
	powervsVersion "github.com/ppc64le-cloud/packer-plugin-powervs/version"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("my-builder", new(powervs.Builder))
	pps.RegisterProvisioner("my-provisioner", new(powervsProv.Provisioner))
	pps.RegisterPostProcessor("my-post-processor", new(powervsPP.PostProcessor))
	pps.RegisterDatasource("my-datasource", new(powervsData.Datasource))
	pps.SetVersion(powervsVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
