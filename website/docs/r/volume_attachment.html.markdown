---
layout: "packet"
page_title: "Packet: packet_volume_attachment"
sidebar_current: "docs-packet-resource-volume-attachment"
description: |-
  Provides attachment of volumes to devices in the Packet Host.
---

# packet\_volume\_attachment

Provides attachment of Packet Block Storage Volume to Devices.

Device and volume must be in the same location (facility).

Once attached by Terraform, they must then be mounted using the `packet-block-storage-attach` and `packet-block-storage-detach` scripts, which are presinstalled on most OS images. They can also be found in [https://github.com/packethost/packet-block-storage](https://github.com/packethost/packet-block-storage).

## Example Usage

Follwing example will create a device, a volume, and then it will attach the volume to the device over the API.

```hcl
resource "packet_device" "test_device_va" {
  hostname         = "terraform-test-device-va"
  plan             = "t1.small.x86"
  facilities       = ["ewr1"]
  operating_system = "ubuntu_16_04"
  billing_cycle    = "hourly"
  project_id       = "${local.project_id}"
}

resource "packet_volume" "test_volume_va" {
  plan = "storage_1"
  billing_cycle = "hourly"
  size = 100
  project_id = "${local.project_id}"
  facility = "ewr1"
  snapshot_policies = { snapshot_frequency = "1day", snapshot_count = 7 }
}

resource "packet_volume_attachment" "test_volume_attachment" {
  device_id = "${packet_device.test_device_va.id}"
  volume_id = "${packet_volume.test_volume_va.id}"
}
```

After applying above hcl, in order to use the volume in the OS of the device, you need to run the attach script. You can run `packet-block-storage-attach` manually over SSH, or you can extend the hcl with following snippet to attach it over remote-exec with Terraform.

```hcl
resource "null_resource" "run_attach_scripts" {
  // re-run the attachment script if any of these resources change
  triggers = {
	device_id = packet_device.test_device_va.id
	volume_id = packet_volume.test_volume_va.id
  }
  connection {
    type     = "ssh"
    user     = "root"
    private_key = file("/home/user/.ssh/id.dsa")
    host     = packet_device.test_device_va.access_public_ipv4
  }
  provisioner "remote-exec" {
    // run the attach script twice for larger chance of success
	inline = [
			"packet-block-storage-attach",
			"packet-block-storage-attach",
    ]
  }
  depends_on = [packet_volume_attachment.test_volume_attachment]
}

```

## Argument Reference

The following arguments are supported:

* `volume_id` - (Required) The ID of the volume to attach
* `device_id` - (Required) The ID of the device to which the volume should be attached

## Attributes Reference

The following attributes are exported:

* `id` - The unique ID of the volume attachment
