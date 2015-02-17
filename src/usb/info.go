// Copyright 2015 Brian Swetland <swetland@frotz.net>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package usb

import "strings"
import "io/ioutil"
import "fmt"

const SYSPATH = "/sys/bus/usb/devices/"

func atou(s []byte) int {
	var n int = 0
	for i := range s {
		if (s[i] < '0') || (s[i] > '9') {
			break
		}
		n = n*10 + int(s[i]) - '0'
	}
	return n
}

type DeviceInfo struct {
	Next   *DeviceInfo
	DevNum int
	BusNum int
	DeviceDescriptor
	Config  []ConfigInfo
	syspath string
	devpath string
}

type ConfigInfo struct {
	ConfigDescriptor
	Interface []InterfaceInfo
}

type InterfaceInfo struct {
	InterfaceDescriptor
	Endpoint []EndpointDescriptor
}

func badDesc(d []byte, kind uint8, size uint8) bool {
	if len(d) < int(size) {
		return true
	}
	if d[0] < size {
		return true
	}
	if len(d) < int(d[0]) {
		return true
	}
	if d[1] != kind {
		return true
	}
	return false
}

func parseDeviceDesc(d []byte, desc *DeviceDescriptor) []byte {
	if badDesc(d, DT_DEVICE, DT_DEVICE_SIZE) {
		return nil
	}
	desc.Length = d[0]
	desc.DescriptorType = d[1]
	desc.UsbVersion = uint16(d[2]) | (uint16(d[3]) << 8)
	desc.DeviceClass = d[4]
	desc.DeviceSubClass = d[5]
	desc.DeviceProtocol = d[6]
	desc.MaxPacketSize0 = d[7]
	desc.VendorID = uint16(d[8]) | (uint16(d[9]) << 8)
	desc.ProductID = uint16(d[10]) | (uint16(d[11]) << 8)
	desc.DeviceVersion = uint16(d[12]) | (uint16(d[13]) << 8)
	desc.ManufacturerIdx = d[14]
	desc.ProductIdx = d[15]
	desc.SerialNumberIdx = d[16]
	desc.NumConfigurations = d[17]
	return d[d[0]:]
}

func parseConfigDesc(d []byte, desc *ConfigDescriptor) []byte {
	if badDesc(d, DT_CONFIG, DT_CONFIG_SIZE) {
		return nil
	}
	desc.Length = d[0]
	desc.DescriptorType = d[1]
	desc.TotalLength = uint16(d[2]) | (uint16(d[3]) << 8)
	desc.NumInterfaces = d[4]
	desc.ConfigurationValue = d[5]
	desc.ConfigurationIdx = d[6]
	desc.Attributes = d[7]
	desc.MaxPower = d[8]
	return d[d[0]:]
}

func parseInterfaceDesc(d []byte, desc *InterfaceDescriptor) []byte {
	if badDesc(d, DT_INTERFACE, DT_INTERFACE_SIZE) {
		return nil
	}
	desc.Length = d[0]
	desc.DescriptorType = d[1]
	desc.InterfaceNumber = d[2]
	desc.AlternateSetting = d[3]
	desc.NumEndpoints = d[4]
	desc.InterfaceClass = d[5]
	desc.InterfaceSubClass = d[6]
	desc.InterfaceProtocol = d[7]
	desc.InterfaceIdx = d[8]
	return d[d[0]:]
}

func parseEndpointDesc(d []byte, desc *EndpointDescriptor) []byte {
	if badDesc(d, DT_ENDPOINT, DT_ENDPOINT_SIZE) {
		return nil
	}
	desc.Length = d[0]
	desc.DescriptorType = d[1]
	desc.EndpointAddress = d[2]
	desc.Attributes = d[3]
	desc.MaxPacketSize = uint16(d[4]) | (uint16(d[5]) << 8)
	desc.Interval = d[6]
	return d[d[0]:]
}

func countDescriptors(d []byte, kind uint8) int {
	count := 0
	for len(d) > 1 {
		if int(d[0]) > len(d) {
			break
		}
		if d[1] == kind {
			count++
		}
		d = d[d[0]:]
	}
	return count
}

func skipNonmatching(d []byte, kind uint8) []byte {
	if len(d) < 2 {
		return d
	}
	if len(d) < int(d[0]) {
		return d
	}
	if d[1] != kind {
		return d[d[0]:]
	}
	return d
}

func parseConfig(d []byte, ci *ConfigInfo) []byte {
	if ci.TotalLength < uint16(ci.Length) {
		return nil
	}
	after := d[ci.TotalLength-uint16(ci.Length):]
	d = d[:ci.TotalLength-uint16(ci.Length)]

	// NumInterfaces does not include alternate settings
	count := countDescriptors(d, DT_INTERFACE)

	ci.Interface = make([]InterfaceInfo, count)
	for i := 0; i < count; i++ {
		d = skipNonmatching(d, DT_INTERFACE)
		d = parseInterfaceDesc(d, &ci.Interface[i].InterfaceDescriptor)
		if d == nil {
			return nil
		}
		ii := &ci.Interface[i]
		ii.Endpoint = make([]EndpointDescriptor, ii.NumEndpoints)
		for j := 0; j < int(ii.NumEndpoints); j++ {
			d = skipNonmatching(d, DT_ENDPOINT)
			d = parseEndpointDesc(d, &ii.Endpoint[j])
			if d == nil {
				return nil
			}
		}
	}
	return after
}

func parseDescriptors(d []byte) *DeviceInfo {
	di := &DeviceInfo{}
	d = parseDeviceDesc(d, &di.DeviceDescriptor)
	if d == nil {
		return nil
	}
	di.Config = make([]ConfigInfo, di.NumConfigurations)
	for i := 0; i < int(di.NumConfigurations); i++ {
		d = skipNonmatching(d, DT_CONFIG)
		d = parseConfigDesc(d, &di.Config[i].ConfigDescriptor)
		if d == nil {
			return nil
		}
		d = parseConfig(d, &di.Config[i])
		if d == nil {
			return nil
		}
	}
	return di
}

func DeviceInfoList() *DeviceInfo {
	var list *DeviceInfo
	var devnum int
	var busnum int
	fi, e := ioutil.ReadDir(SYSPATH)
	if e != nil {
		return nil
	}
	for i := range fi {
		if strings.IndexByte(fi[i].Name(), ':') != -1 {
			continue
		}
		s, e := ioutil.ReadFile(SYSPATH + fi[i].Name() + "/devnum")
		if e != nil {
			continue
		}
		devnum = atou(s)
		s, e = ioutil.ReadFile(SYSPATH + fi[i].Name() + "/busnum")
		if e != nil {
			continue
		}
		busnum = atou(s)
		desc, e := ioutil.ReadFile(SYSPATH + fi[i].Name() + "/descriptors")
		if e != nil {
			continue
		}
		di := parseDescriptors(desc)
		if di != nil {
			di.BusNum = busnum
			di.DevNum = devnum
			di.syspath = SYSPATH + fi[i].Name()
			di.devpath = fmt.Sprintf("/dev/bus/usb/%03d/%03d", busnum, devnum)
			di.Next = list
			list = di
		}
	}
	return list
}
