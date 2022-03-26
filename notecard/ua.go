// Copyright 2017 Blues Inc.  All rights reserved.
// Use of this source code is governed by licenses granted by the
// copyright holder including that found in the LICENSE file.

package notecard

import (
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	/*	"github.com/shirou/gopsutil/v3/host"	// Deprecated */
	"github.com/shirou/gopsutil/v3/mem"
)

// UserAgent generates a User Agent object for a given interface
func (context *Context) UserAgent() (ua map[string]interface{}) {

	ua = map[string]interface{}{}
	ua["agent"] = "note-go"
	ua["compiler"] = fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	ua["req_interface"] = context.iface
	if context.isSerial {
		ua["req_port"] = context.serialName
	} else {
		ua["req_port"] = context.port
	}

	m, _ := mem.VirtualMemory()
	ua["cpu_mem"] = m.Total

	c, _ := cpu.Info()
	if len(c) >= 1 {
		ua["cpu_mhz"] = int(c[0].Mhz)
		ua["cpu_cores"] = int(c[0].Cores)
		ua["cpu_vendor"] = c[0].VendorID
		ua["cpu_name"] = c[0].ModelName
	}

	/* Deprecated
	h, _ := host.Info()
	ua["os_name"] = h.OS                 // freebsd, linux
	ua["os_platform"] = h.Platform       // ubuntu, linuxmint
	ua["os_family"] = h.PlatformFamily   // debian, rhel
	ua["os_version"] = h.PlatformVersion // complete OS version
	*/

	return

}
