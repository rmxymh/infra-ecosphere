package main

import (
	"net"
	"log"
	"bytes"
	"encoding/binary"
	"github.com/rmxymh/infra-ecosphere/utils"
	"github.com/rmxymh/infra-ecosphere/ipmi"
	"fmt"
	"github.com/rmxymh/infra-ecosphere/web"
	"github.com/jmcvetta/napping"
)

var EcosphereIP string = "10.0.2.2"
var EcospherePort int = 9090

func SetBootDevice(addr *net.UDPAddr, server *net.UDPConn, wrapper ipmi.IPMISessionWrapper, message ipmi.IPMIMessage, selector ipmi.IPMIChassisBootOptionParameterSelector) {
	localIP := utils.GetLocalIP(server)

	buf := bytes.NewBuffer(selector.Parameters)
	request := ipmi.IPMIChassisBootOptionBootFlags{}
	binary.Read(buf, binary.BigEndian, &request)

	// Simulate: We just dump log but do nothing here.
	if request.BootParam & ipmi.BOOT_PARAM_BITMASK_VALID != 0 {
		log.Println("        IPMI CHASSIS BOOT FLAG: Valid")
	}
	if request.BootParam & ipmi.BOOT_PARAM_BITMASK_PERSISTENT != 0 {
		log.Println("        IPMI CHASSIS BOOT FLAG: Persistent")
	} else {
		log.Println("        IPMI CHASSIS BOOT FLAG: Only on the next boot")
	}
	if request.BootParam & ipmi.BOOT_PARAM_BITMASK_BOOT_TYPE_EFI != 0 {
		log.Println("        IPMI CHASSIS BOOT FLAG: Boot Type = EFI")
	} else {
		log.Println("        IPMI CHASSIS BOOT FLAG: Boot Type = PC Compatible (Legacy)")
	}

	// Simulate: We just dump log but do nothing here
	if request.BootDevice & ipmi.BOOT_DEVICE_BITMASK_CMOS_CLEAR != 0 {
		log.Println("        IPMI CHASSIS BOOT DEVICE: CMOS Clear")
	}
	if request.BootDevice & ipmi.BOOT_DEVICE_BITMASK_LOCK_KEYBOARD != 0 {
		log.Println("        IPMI CHASSIS BOOT DEVICE: Lock Keyboard")
	}
	if request.BootDevice & ipmi.BOOT_DEVICE_BITMASK_SCREEN_BLANK != 0 {
		log.Println("        IPMI CHASSIS BOOT DEVICE: Screen Blank")
	}
	if request.BootDevice & ipmi.BOOT_DEVICE_BITMASK_LOCK_RESET != 0 {
		log.Println("        IPMI CHASSIS BOOT DEVICE: Lock RESET Buttons")
	}

	// This part contains some options that we only support: PXE, CD, HDD
	//   Maybe there is another way to simulate remote device.
	device := (request.BootDevice & ipmi.BOOT_DEVICE_BITMASK_DEVICE) >> 2

	bootdevReq := web.WebReqBootDev{}
	bootdevResp := web.WebRespBootDev{}
	baseAPI := fmt.Sprintf("http://%s:%d/api/BMCs/%s/bootdev", EcosphereIP, EcospherePort, localIP)

	switch device {
	case ipmi.BOOT_DEVICE_FORCE_PXE:
		log.Println("        IPMI CHASSIS BOOT DEVICE: BOOT_DEVICE_FORCE_PXE")
		bootdevReq.Device = "PXE"
		resp, err := napping.Put(baseAPI, &bootdevReq, &bootdevResp, nil)
		if err != nil {
			log.Println("Failed to call ecophsere Web API for setting bootdev: ", err.Error())
		} else if resp.Status() != 200 {
			log.Println("Failed to call ecosphere Web API for setting bootdev: ", bootdevResp.Status)
		}

	case ipmi.BOOT_DEVICE_FORCE_HDD:
		log.Println("        IPMI CHASSIS BOOT DEVICE: BOOT_DEVICE_FORCE_HDD")
		bootdevReq.Device = "DISK"
		resp, err := napping.Put(baseAPI, &bootdevReq, &bootdevResp, nil)
		if err != nil {
			log.Println("Failed to call ecophsere Web API for setting bootdev: ", err.Error())
		} else if resp.Status() != 200 {
			log.Println("Failed to call ecosphere Web API for setting bootdev: ", bootdevResp.Status)
		}

	case ipmi.BOOT_DEVICE_FORCE_HDD_SAFE:
		log.Println("        IPMI CHASSIS BOOT DEVICE: BOOT_DEVICE_FORCE_HDD_SAFE")
	case ipmi.BOOT_DEVICE_FORCE_DIAG_PARTITION:
		log.Println("        IPMI CHASSIS BOOT DEVICE: BOOT_DEVICE_FORCE_DIAG_PARTITION")
	case ipmi.BOOT_DEVICE_FORCE_CD:
		log.Println("        IPMI CHASSIS BOOT DEVICE: BOOT_DEVICE_FORCE_CD")
	case ipmi.BOOT_DEVICE_FORCE_BIOS:
		log.Println("        IPMI CHASSIS BOOT DEVICE: BOOT_DEVICE_FORCE_BIOS")
	case ipmi.BOOT_DEVICE_FORCE_REMOTE_FLOPPY:
		log.Println("        IPMI CHASSIS BOOT DEVICE: BOOT_DEVICE_FORCE_REMOTE_FLOPPY")
	case ipmi.BOOT_DEVICE_FORCE_REMOTE_MEDIA:
		log.Println("        IPMI CHASSIS BOOT DEVICE: BOOT_DEVICE_FORCE_REMOTE_MEDIA")
	case ipmi.BOOT_DEVICE_FORCE_REMOTE_CD:
		log.Println("        IPMI CHASSIS BOOT DEVICE: BOOT_DEVICE_FORCE_REMOTE_CD")
	case ipmi.BOOT_DEVICE_FORCE_REMOTE_HDD:
		log.Println("        IPMI CHASSIS BOOT DEVICE: BOOT_DEVICE_FORCE_REMOTE_HDD")
	}

	// Simulate: We just dump log but do nothing here.
	if request.BIOSVerbosity & ipmi.BOOT_BIOS_BITMASK_LOCK_VIA_POWER != 0 {
		log.Println("        IPMI CHASSIS BOOT DEVICE: Lock out (power off / sleep request) via Power Button")
	}
	if request.BIOSVerbosity & ipmi.BOOT_BIOS_BITMASK_EVENT_TRAP != 0 {
		log.Println("        IPMI CHASSIS BOOT DEVICE: Force Progress Event Trap (Only for IPMI 2.0)")
	}
	if request.BIOSVerbosity & ipmi.BOOT_BIOS_BITMASK_PASSWORD_BYPASS != 0 {
		log.Println("        IPMI CHASSIS BOOT DEVICE: User password bypass")
	}
	if request.BIOSVerbosity & ipmi.BOOT_BIOS_BITMASK_LOCK_SLEEP != 0 {
		log.Println("        IPMI CHASSIS BOOT DEVICE: Lock out Sleep Button")
	}
	verbosity := (request.BIOSVerbosity & ipmi.BOOT_BIOS_BITMASK_FIRMWARE) >> 5
	switch verbosity {
	case ipmi.BOOT_BIOS_FIRMWARE_SYSTEM_DEFAULT:
		log.Println("        IPMI CHASSIS BOOT BIOS: BOOT_BIOS_FIRMWARE_SYSTEM_DEFAULT")
	case ipmi.BOOT_BIOS_FIRMWARE_REQUEST_QUIET:
		log.Println("        IPMI CHASSIS BOOT BIOS: BOOT_BIOS_FIRMWARE_REQUEST_QUIET")
	case ipmi.BOOT_BIOS_FIRMWARE_REQUEST_VERBOSE:
		log.Println("        IPMI CHASSIS BOOT BIOS: BOOT_BIOS_FIRMWARE_REQUEST_VERBOSE")
	}
	console_redirect := (request.BIOSVerbosity & ipmi.BOOT_BIOS_BITMASK_CONSOLE_REDIRECT)
	switch console_redirect {
	case ipmi.BOOT_BIOS_CONSOLE_REDIRECT_OCCURS_PER_BIOS_SETTING:
		log.Println("        IPMI CHASSIS BOOT BIOS: BOOT_BIOS_CONSOLE_REDIRECT_OCCURS_PER_BIOS_SETTING")
	case ipmi.BOOT_BIOS_CONSOLE_REDIRECT_SUPRESS_CONSOLE_IF_ENABLED:
		log.Println("        IPMI CHASSIS BOOT BIOS: BOOT_BIOS_CONSOLE_REDIRECT_SUPRESS_CONSOLE_IF_ENABLED")
	case ipmi.BOOT_BIOS_CONSOLE_REDIRECT_REQUEST_ENABLED:
		log.Println("        IPMI CHASSIS BOOT BIOS: BOOT_BIOS_CONSOLE_REDIRECT_REQUEST_ENABLED")
	}

	// Simulate: We just dump log but do nothing here.
	if request.BIOSSharedMode & ipmi.BOOT_BIOS_SHARED_BITMASK_OVERRIDE != 0 {
		log.Println("        IPMI CHASSIS BOOT BIOS: BOOT_BIOS_SHARED_BITMASK_OVERRIDE")
	}
	mux_control := request.BIOSSharedMode & ipmi.BOOT_BIOS_SHARED_BITMASK_MUX_CONTROL_OVERRIDE
	switch mux_control {
	case ipmi.BOOT_BIOS_SHARED_MUX_RECOMMENDED:
		log.Println("        IPMI CHASSIS BOOT BIOS: BOOT_BIOS_SHARED_MUX_RECOMMENDED")
	case ipmi.BOOT_BIOS_SHARED_MUX_TO_SYSTEM:
		log.Println("        IPMI CHASSIS BOOT BIOS: BOOT_BIOS_SHARED_MUX_TO_SYSTEM")
	case ipmi.BOOT_BIOS_SHARED_MUX_TO_BMC:
		log.Println("        IPMI CHASSIS BOOT BIOS: BOOT_BIOS_SHARED_MUX_TO_BMC")
	}

	ipmi.SendIPMIChassisSetBootOptionResponseBack(addr, server, wrapper, message);
}
