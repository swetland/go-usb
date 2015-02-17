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

const (
	// descriptor types
	DT_DEVICE             = 0x01
	DT_CONFIG             = 0x02
	DT_STRING             = 0x03
	DT_INTERFACE          = 0x04
	DT_ENDPOINT           = 0x05
	DT_DEVICE_QUALIFIER   = 0x06
	DT_OTHER_SPEED_CONFIG = 0x07
	DT_INTERFACE_POWER    = 0x08

	// descriptor sizes
	DT_DEVICE_SIZE         = 18
	DT_CONFIG_SIZE         = 9
	DT_INTERFACE_SIZE      = 9
	DT_ENDPOINT_SIZE       = 7
	DT_ENDPOINT_AUDIO_SIZE = 9

	// endpoint address
	ENDPOINT_IN = 0x80

	// endpoint attributes
	ENDPOINT_XFER_CONTROL = 0
	ENDPOINT_XFER_ISOC    = 1
	ENDPOINT_XFER_BULK    = 2
	ENDPOINT_XFER_INT     = 3
	ENDPOINT_XFER_MASK    = 3
)

type DeviceDescriptor struct {
	Length            uint8
	DescriptorType    uint8
	UsbVersion        uint16 // bcd
	DeviceClass       uint8
	DeviceSubClass    uint8
	DeviceProtocol    uint8
	MaxPacketSize0    uint8
	VendorID          uint16
	ProductID         uint16
	DeviceVersion     uint16 // bcd
	ManufacturerIdx   uint8
	ProductIdx        uint8
	SerialNumberIdx   uint8
	NumConfigurations uint8
}

type ConfigDescriptor struct {
	Length             uint8
	DescriptorType     uint8
	TotalLength        uint16
	NumInterfaces      uint8
	ConfigurationValue uint8
	ConfigurationIdx   uint8
	Attributes         uint8
	MaxPower           uint8
}

type InterfaceDescriptor struct {
	Length            uint8
	DescriptorType    uint8
	InterfaceNumber   uint8
	AlternateSetting  uint8
	NumEndpoints      uint8
	InterfaceClass    uint8
	InterfaceSubClass uint8
	InterfaceProtocol uint8
	InterfaceIdx      uint8
}

type EndpointDescriptor struct {
	Length          uint8
	DescriptorType  uint8
	EndpointAddress uint8
	Attributes      uint8
	MaxPacketSize   uint16
	Interval        uint8
}

type ControlRequest struct {
	RequestType uint8
	Request     uint8
	Value       uint16
	Index       uint16
	Length      uint16
}
