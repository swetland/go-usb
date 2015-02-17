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

import "syscall"
import "os"
import "unsafe"
import "sync"
import "log"

type Transfer struct {
	Status int32          // transaction status (0 == success)
	Length int32          // length of data transferred
	Data   []byte         // data to transmit or receive
	Done   chan *Transfer // written to on completion
	urb    usbdevfs_urb
}

type Device struct {
	fd     int
	lock   sync.Mutex
	active map[uintptr]*Transfer
	log    *log.Logger
}

func (u *Device) queue(kind uint8, ep uint8, xfer *Transfer) error {
	if xfer.Data == nil {
		return syscall.EINVAL
	}
	urb := &xfer.urb
	urb.urbtype = kind
	urb.endpoint = ep
	urb.status = 0
	urb.flags = 0
	urb.buffer = uintptr(unsafe.Pointer(&xfer.Data[0]))
	urb.buffer_length = int32(len(xfer.Data))
	urb.actual_length = 0
	urb.start_frame = 0
	urb.number_of_packets = 0
	urb.error_count = 0
	urb.signr = 0
	urb.usercontext = 0
	n := uintptr(unsafe.Pointer(urb))
	u.lock.Lock()
	u.active[n] = xfer
	u.lock.Unlock()
	_, e := ioctl(u.fd, USBDEVFS_SUBMITURB, n)
	if e != nil {
		u.lock.Lock()
		delete(u.active, n)
		u.lock.Unlock()
	}
	return e
}

func (u *Device) QueueBulk(ep uint8, xfer *Transfer) error {
	return u.queue(URB_TYPE_BULK, ep, xfer)
}

func (u *Device) QueueInterrupt(ep uint8, xfer *Transfer) error {
	return u.queue(URB_TYPE_INTERRUPT, ep, xfer)
}

func (u *Device) QueueControl(ep uint8, xfer *Transfer) error {
	return u.queue(URB_TYPE_CONTROL, ep, xfer)
}

// This ioctl is interruptible by signals and will not wedge the process on
// exit, but otherwise blocks forever if no URBs complete.  Probably should
// poll/select and use the nonblocking version to allow for better cleanup
// on Close()
func (u *Device) reaper() {
	for {
		var n uintptr
		_, e := ioctl(u.fd, USBDEVFS_REAPURB, uintptr(unsafe.Pointer(&n)))
		if e != nil {
			u.log.Println("failure reaping URBs", e)
			break
		}
		u.lock.Lock()
		xfer := u.active[n]
		delete(u.active, n)
		u.lock.Unlock()
		if xfer == nil {
			u.log.Println("kernel returned invalid urb pointer?!")
			continue
		}
		xfer.Status = xfer.urb.status
		xfer.Length = xfer.urb.actual_length
		u.log.Println("status ", xfer.urb.status)
		u.log.Println("actual ", xfer.urb.actual_length)
		if xfer.Done != nil {
			xfer.Done <- xfer
		}
	}
}

func OpenVidPid(vid uint16, pid uint16) (*Device, error) {
	for di := DeviceInfoList(); di != nil; di = di.Next {
		if (vid != di.VendorID) || (pid != di.ProductID) {
			continue
		}
		return Open(di)
	}
	return nil, syscall.ENODEV
}

func OpenBusDev(bus int, dev int) (*Device, error) {
	for di := DeviceInfoList(); di != nil; di = di.Next {
		if (bus != di.BusNum) || (dev != di.DevNum) {
			continue
		}
		return Open(di)
	}
	return nil, syscall.ENODEV
}

func Open(di *DeviceInfo) (*Device, error) {
	fd, e := syscall.Open(di.devpath, os.O_RDWR|syscall.O_CLOEXEC, 0666)
	if e != nil {
		return nil, e
	}
	dev := &Device{
		fd:     fd,
		active: make(map[uintptr]*Transfer),
		log:    log.New(os.Stderr, "usb: ", 0),
	}
	go dev.reaper()
	return dev, nil
}

func (u *Device) Close() {
	// TODO: sanely shutdown reaper
	u.lock.Lock()
	syscall.Close(u.fd)
	u.fd = -1
	u.lock.Unlock()
}

func (u *Device) ClaimInterface(n uint32) error {
	_, e := ioctl(u.fd, USBDEVFS_CLAIMINTERFACE, uintptr(unsafe.Pointer(&n)))
	return e
}

func (u *Device) ControlTransfer(
	reqtype uint8, request uint8, value uint16, index uint16,
	length uint16, timeout uint32, data []byte) (int, error) {

	if int(length) > len(data) {
		return 0, syscall.ENOSPC
	}
	p := unsafe.Pointer(&data[0])
	ct := ctrltransfer{reqtype, request, value, index, length, timeout, 0, uintptr(p)}
	n, e := ioctl(u.fd, USBDEVFS_CONTROL, uintptr(unsafe.Pointer(&ct)))
	return n, e
}

func (u *Device) BulkTransfer(
	endpoint uint32, length uint32, timeout uint32, data []byte) (int, error) {

	if int(length) > len(data) {
		return 0, syscall.ENOSPC
	}
	p := unsafe.Pointer(&data[0])
	bt := bulktransfer{endpoint, length, timeout, 0, uintptr(p)}
	n, e := ioctl(u.fd, USBDEVFS_BULK, uintptr(unsafe.Pointer(&bt)))
	return n, e
}

func ioctl(fd int, req uintptr, arg uintptr) (int, error) {
	r, _, e := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), req, arg)
	if e == 0 {
		return int(r), nil
	} else {
		return 0, e
	}
}
