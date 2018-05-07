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

const defaultUbuntuImageID = "ubuntu_16_0402_64_20G_alibase_20171227.vhd"
const defaultUbuntuImagePrefix = "ubuntu_16_0402_64"

func validateECSRegion(region string) (common.Region, error) {
	for _, v := range common.ValidRegions {
		if v == common.Region(region) {
			return v, nil
		}
	}

	return "", errInvalidRegion
}

const digitals = "0123456789"
const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const specialChars = "()`~!@#$%^&*-+=|{}[]:;'<>,.?/"
const dictionary = digitals + alphabet + specialChars
const paswordLen = 16

func randomPassword() string {
	var bytes = make([]byte, paswordLen)
	rand.Read(bytes)
	for k, v := range bytes {
		var ch byte

		switch k {
		case 0:
			ch = alphabet[v%byte(len(alphabet))]
		case 1:
			ch = digitals[v%byte(len(digitals))]
		case 2:
			ch = specialChars[v%byte(len(specialChars))]
		default:
			ch = dictionary[v%byte(len(dictionary))]
		}
		bytes[k] = ch
	}
	return string(bytes)
}

func isUbuntuImage(image string) bool {
	return strings.HasPrefix(image, "ubuntu")
}
