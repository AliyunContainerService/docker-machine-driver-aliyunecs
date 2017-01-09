package aliyunecs

import "testing"

func testIsUbuntuImage(t *testing.T) {

	if !isUbuntuImage("ubuntu_160401_64_40G_cloudinit_20161115.vhd") {
		t.Fatalf("Failed to check ubuntu image - ubuntu 16.04")
	}

	if !isUbuntuImage("ubuntu_140405_64_40G_cloudinit_20161115.vhd") {
		t.Fatalf("Failed to check ubuntu image - ubuntu 14.04")
	}

	if isUbuntuImage("centos-test.vhd") {
		t.Fatalf("Failed to check ubuntu image - centos")
	}
}
