# Use case: Testing packer-plugin-powervs with [image builder](https://github.com/kubernetes-sigs/image-builder)

Note: The following steps are being executed on x86 VM.

1. Clone packer-plugin-powervs
   ```shell
    $ git clone https://github.com/ppc64le-cloud/packer-plugin-powervs.git
   ```
2. Build the binary
    ```
   $ cd packer-plugin-powervs
   $ make build
   
   This will build binary named `packer-plugin-powervs` under current directoy.
   ```
3. Clone the image-builder
    ```
   $ git clone https://github.com/kubernetes-sigs/image-builder.git
   ```
4. Install the necessary dependencies for image-builder
    ```
   $ cd images/capi
   $ make deps-powervs
   ```
   If the above command returns packages required error manually install them and retry.



5. Replace the downloaded PowerVS packer plugin with binary built in step 2

   1. When we run ```make deps-powervs```, It will [download and install](https://github.com/kubernetes-sigs/image-builder/blob/main/images/capi/hack/ensure-powervs.sh#L42-L52) the PowerVS plugin in `"${HOME}/.packer.d/plugins/packer-plugin-powervs"` directory.
   2. Get the name of the downloaded binary and replace it with binary built in step 2.
   ```shell
   $ ls ${HOME}/.packer.d/plugins/packer-plugin-powervs/
   $ mv /root/packer-plugin-powervs/packer-plugin-powervs ${HOME}/.packer.d/plugins/packer-plugin-powervs/packer-plugin-powervs_v0.2.1_x5.0_linux_amd64
   ```
   
6. Update the image-builder script to avoid downloading the packer-plugin-powervs binary again from upstream.
   1. When we run step 7, packer-plugin-powervs binary will be downloaded again from upstream, which will replace the binary we added in step 5.
   2. To avoid replacing of binary,edit the [ensure-powervs.sh](https://github.com/kubernetes-sigs/image-builder/blob/main/images/capi/hack/ensure-powervs.sh#L42-L52) file to remove download and moving of upstream binary.
   
7. Run the image builder, Instructions can be found in [cluster-api-provider-ibmcloud](https://cluster-api-ibmcloud.sigs.k8s.io/developer/build-images#powervs)