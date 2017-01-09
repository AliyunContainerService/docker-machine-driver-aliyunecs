package aliyunecs

import (
	"crypto/rand"
	"errors"
	"strings"

	"github.com/denverdino/aliyungo/common"
)

var (
	errInvalidRegion  = errors.New("invalid region specified")
	errNoVpcs         = errors.New("No VPCs found in region")
	errMachineFailure = errors.New("Machine failed to start")
	errNoIP           = errors.New("No IP Address associated with the instance")
	errComplete       = errors.New("Complete")
)

const defaultUbuntuImageID = "ubuntu_160401_64_40G_cloudinit_20161115.vhd"
const defaultUbuntuImagePrefix = "ubuntu_160401_64_40G_"

func validateECSRegion(region string) (common.Region, error) {
	for _, v := range common.ValidRegions {
		if v == common.Region(region) {
			return v, nil
		}
	}

	return "", errInvalidRegion
}

const dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const paswordLen = 16

func randomPassword() string {
	var bytes = make([]byte, paswordLen)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}

func isUbuntuImage(image string) bool {
	return strings.HasPrefix(image, "ubuntu")
}
