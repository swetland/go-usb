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
	USBDEVFS_CONTROL          = 0xc0185500
	USBDEVFS_BULK             = 0xc0185502
	USBDEVFS_RESETEP          = 0x80045503
	USBDEVFS_SETINTERFACE     = 0x80085504
	USBDEVFS_SETCONFIGURATION = 0x80045505
	USBDEVFS_GETDRIVER        = 0x41045508
	USBDEVFS_SUBMITURB        = 0x8038550a
	USBDEVFS_DISCARDURB       = 0x0000550b
	USBDEVFS_REAPURB          = 0x4008550c
	USBDEVFS_REAPURBNDELAY    = 0x4008550d
	USBDEVFS_DISCSIGNAL       = 0x8010550e
	USBDEVFS_CLAIMINTERFACE   = 0x8004550f
	USBDEVFS_RELEASEINTERFACE = 0x80045510
	USBDEVFS_CONNECTINFO      = 0x40085511
	USBDEVFS_IOCTL            = 0xc0105512
	USBDEVFS_HUB_PORTINFO     = 0x80805513
	USBDEVFS_RESET            = 0x00005514
	USBDEVFS_CLEAR_HALT       = 0x80045515
	USBDEVFS_DISCONNECT       = 0x00005516
	USBDEVFS_CONNECT          = 0x00005517
	USBDEVFS_CLAIM_PORT       = 0x80045518
	USBDEVFS_RELEASE_PORT     = 0x80045519
	USBDEVFS_GET_CAPABILITIES = 0x8004551a
	USBDEVFS_DISCONNECT_CLAIM = 0x8108551b
)

type ctrltransfer struct {
	bRequestType uint8
	bRequest     uint8
	wValue       uint16
	wIndex       uint16
	wLength      uint16
	timeout      uint32 // ms
	_pad0        uint32
	data         uintptr
}

type bulktransfer struct {
	endpoint uint32
	length   uint32
	timeout  uint32 // ms
	_pad0    uint32
	data     uintptr
}

const (
	URB_TYPE_ISO       = 0
	URB_TYPE_INTERRUPT = 1
	URB_TYPE_CONTROL   = 2
	URB_TYPE_BULK      = 3
)

const (
	URB_FLAG_SHORT_NOT_OK      = 0x01
	URB_FLAG_ISO_ASAP          = 0x02
	URB_FLAG_BULK_CONTINUATION = 0x04
	URB_FLAG_NO_FSBR           = 0x20
	URB_FLAG_ZERO_PACKET       = 0x40
	URB_FLAG_NO_INTERRUPT      = 0x80
)

type usbdevfs_urb struct {
	urbtype           uint8
	endpoint          uint8
	_pad0             uint8
	_pad1             uint8
	status            int32
	flags             uint32
	_pad2             uint32
	buffer            uintptr
	buffer_length     int32
	actual_length     int32
	start_frame       int32
	number_of_packets int32
	error_count       int32
	signr             uint32
	usercontext       uintptr
}

type usbdevfs_setifc struct {
	num uint32
	alt uint32
}

type usbdevfs_ioctl struct {
	ifc  uint32
	code uint32
	data uintptr
}
