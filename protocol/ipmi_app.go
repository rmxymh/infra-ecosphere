package protocol

import (
	"net"
	"bytes"
	"log"
	"encoding/binary"
	"fmt"
)

// port from OpenIPMI
// App Network Function
const (
	IPMI_CMD_GET_DEVICE_ID = 			0x01
	IPMI_CMD_BROADCAST_GET_DEVICE_ID = 		0x01
	IPMI_CMD_COLD_RESET = 				0x02
	IPMI_CMD_WARM_RESET = 				0x03
	IPMI_CMD_GET_SELF_TEST_RESULTS = 		0x04
	IPMI_CMD_MANUFACTURING_TEST_ON = 		0x05
	IPMI_CMD_SET_ACPI_POWER_STATE = 		0x06
	IPMI_CMD_GET_ACPI_POWER_STATE = 		0x07
	IPMI_CMD_GET_DEVICE_GUID = 			0x08
	IPMI_CMD_RESET_WATCHDOG_TIMER = 		0x22
	IPMI_CMD_SET_WATCHDOG_TIMER = 			0x24
	IPMI_CMD_GET_WATCHDOG_TIMER = 			0x25
	IPMI_CMD_SET_BMC_GLOBAL_ENABLES = 		0x2e
	IPMI_CMD_GET_BMC_GLOBAL_ENABLES = 		0x2f
	IPMI_CMD_CLEAR_MSG_FLAGS = 			0x30
	IPMI_CMD_GET_MSG_FLAGS = 			0x31
	IPMI_CMD_ENABLE_MESSAGE_CHANNEL_RCV = 		0x32
	IPMI_CMD_GET_MSG = 				0x33
	IPMI_CMD_SEND_MSG = 				0x34
	IPMI_CMD_READ_EVENT_MSG_BUFFER = 		0x35
	IPMI_CMD_GET_BT_INTERFACE_CAPABILITIES = 	0x36
	IPMI_CMD_GET_SYSTEM_GUID = 			0x37
	IPMI_CMD_GET_CHANNEL_AUTH_CAPABILITIES = 	0x38
	IPMI_CMD_GET_SESSION_CHALLENGE = 		0x39
	IPMI_CMD_ACTIVATE_SESSION = 			0x3a
	IPMI_CMD_SET_SESSION_PRIVILEGE = 		0x3b
	IPMI_CMD_CLOSE_SESSION = 			0x3c
	IPMI_CMD_GET_SESSION_INFO = 			0x3d

	IPMI_CMD_GET_AUTHCODE = 			0x3f
	IPMI_CMD_SET_CHANNEL_ACCESS = 			0x40
	IPMI_CMD_GET_CHANNEL_ACCESS = 			0x41
	IPMI_CMD_GET_CHANNEL_INFO = 			0x42
	IPMI_CMD_SET_USER_ACCESS = 			0x43
	IPMI_CMD_GET_USER_ACCESS = 			0x44
	IPMI_CMD_SET_USER_NAME = 			0x45
	IPMI_CMD_GET_USER_NAME = 			0x46
	IPMI_CMD_SET_USER_PASSWORD = 			0x47
	IPMI_CMD_ACTIVATE_PAYLOAD = 			0x48
	IPMI_CMD_DEACTIVATE_PAYLOAD = 			0x49
	IPMI_CMD_GET_PAYLOAD_ACTIVATION_STATUS = 	0x4a
	IPMI_CMD_GET_PAYLOAD_INSTANCE_INFO = 		0x4b
	IPMI_CMD_SET_USER_PAYLOAD_ACCESS = 		0x4c
	IPMI_CMD_GET_USER_PAYLOAD_ACCESS = 		0x4d
	IPMI_CMD_GET_CHANNEL_PAYLOAD_SUPPORT = 		0x4e
	IPMI_CMD_GET_CHANNEL_PAYLOAD_VERSION = 		0x4f
	IPMI_CMD_GET_CHANNEL_OEM_PAYLOAD_INFO = 	0x50

	IPMI_CMD_MASTER_READ_WRITE = 			0x52

	IPMI_CMD_GET_CHANNEL_CIPHER_SUITES = 		0x54
	IPMI_CMD_SUSPEND_RESUME_PAYLOAD_ENCRYPTION = 	0x55
	IPMI_CMD_SET_CHANNEL_SECURITY_KEY = 		0x56
	IPMI_CMD_GET_SYSTEM_INTERFACE_CAPABILITIES = 	0x57
)

const (
	AUTH_NONE = 		0x01
	AUTH_MD2 = 		0x02
	AUTH_MD5 = 		0x04
	AUTH_STRAIGHT_KEY =	0x10
	AUTH_OEM = 		0x20
	AUTH_IPMI_V2 = 		0x80
)

const (
	AUTH_STATUS_ANONYMOUS =		0x01
	AUTH_STATUS_NULL_USER =		0x02
	AUTH_STATUS_NON_NULL_USER =	0x04
	AUTH_STATUS_USER_LEVEL = 	0x08
	AUTH_STATUS_PER_MESSAGE = 	0x10
	AUTH_STATUS_KG =		0x20
)

type IPMIAuthenticationCapabilitiesRequest struct {
	AutnticationTypeSupport uint8
	RequestedPrivilegeLevel uint8
}

type IPMIAuthenticationCapabilitiesResponse struct {
	Channel uint8
	AuthenticationTypeSupport uint8
	AuthenticationStatus uint8
	ExtCapabilities uint8			// In IPMI v1.5, 0 is always put here. (Reserved)
	OEMID [3]uint8
	OEMAuxiliaryData uint8

}

const (
	LEN_IPMI_AUTH_CAPABILITIES_RESPONSE = 8
)

func dumpByteBuffer(buf bytes.Buffer) {
	bytebuf := buf.Bytes()
	fmt.Print("[")
	for i := 0 ; i < buf.Len(); i += 1 {
		fmt.Printf(" %02x", bytebuf[i])
	}
	fmt.Println("]")
}

func HandleIPMIAuthenticationCapabilities(addr *net.UDPAddr, server *net.UDPConn, wrapper IPMISessionWrapper, message IPMIMessage) {
	buf := bytes.NewBuffer(message.Data)
	request := IPMIAuthenticationCapabilitiesRequest{}

	binary.Read(buf, binary.BigEndian, &request)

	// We don't simulate OEM related behavior
	response := IPMIAuthenticationCapabilitiesResponse{}
	response.Channel = 1
	response.AuthenticationTypeSupport = AUTH_MD5 | AUTH_MD2 | AUTH_NONE
	response.AuthenticationStatus = AUTH_STATUS_NON_NULL_USER | AUTH_STATUS_NULL_USER
	response.ExtCapabilities = 0
	response.OEMAuxiliaryData = 0

	dataBuf := bytes.Buffer{}
	binary.Write(&dataBuf, binary.BigEndian, response)

	responseMessage := IPMIMessage{}
	responseMessage.TargetAddress = message.SourceAddress
	remoteLun := message.SourceLun & 0x03
	localLun := message.TargetLun & 0x03
	responseMessage.TargetLun = remoteLun | ((IPMI_NETFN_APP | IPMI_NETFN_RESPONSE) << 2)
	responseMessage.SourceAddress = message.TargetAddress
	responseMessage.SourceLun = localLun
	responseMessage.Command = IPMI_CMD_GET_CHANNEL_AUTH_CAPABILITIES
	responseMessage.CompletionCode = 0
	responseMessage.Data = dataBuf.Bytes()


	responseWrapper := IPMISessionWrapper{}
	responseWrapper.AuthenticationType = wrapper.AuthenticationType
	responseWrapper.SequenceNumber = 0x00
	responseWrapper.SessionId = 0x00

	rmcp := BuildUpRMCPForIPMI()

	obuf := bytes.Buffer{}
	SerializeRMCP(&obuf, rmcp)
	SerializeIPMI(&obuf, responseWrapper, responseMessage)

	dumpByteBuffer(obuf)

	server.WriteToUDP(obuf.Bytes(), addr)
}


func IPMI_APP_DeserializeAndExecute(addr *net.UDPAddr, server *net.UDPConn, wrapper IPMISessionWrapper, message IPMIMessage) {
	switch message.Command {
	case IPMI_CMD_GET_DEVICE_ID:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_DEVICE_ID")
	case IPMI_CMD_COLD_RESET:
		log.Println("      IPMI APP: Command = IPMI_CMD_COLD_RESET")
	case IPMI_CMD_WARM_RESET:
		log.Println("      IPMI APP: Command = IPMI_CMD_WARM_RESET")
	case IPMI_CMD_GET_SELF_TEST_RESULTS:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_SELF_TEST_RESULTS")
	case IPMI_CMD_MANUFACTURING_TEST_ON:
		log.Println("      IPMI APP: Command = IPMI_CMD_MANUFACTURING_TEST_ON")
	case IPMI_CMD_SET_ACPI_POWER_STATE:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_ACPI_POWER_STATE")
	case IPMI_CMD_GET_ACPI_POWER_STATE:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_ACPI_POWER_STATE")
	case IPMI_CMD_GET_DEVICE_GUID:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_DEVICE_GUID")
	case IPMI_CMD_RESET_WATCHDOG_TIMER:
		log.Println("      IPMI APP: Command = IPMI_CMD_RESET_WATCHDOG_TIMER")
	case IPMI_CMD_SET_WATCHDOG_TIMER:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_WATCHDOG_TIMER")
	case IPMI_CMD_GET_WATCHDOG_TIMER:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_WATCHDOG_TIMER")
	case IPMI_CMD_SET_BMC_GLOBAL_ENABLES:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_BMC_GLOBAL_ENABLES")
	case IPMI_CMD_GET_BMC_GLOBAL_ENABLES:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_BMC_GLOBAL_ENABLES")
	case IPMI_CMD_CLEAR_MSG_FLAGS:
		log.Println("      IPMI APP: Command =IPMI_CMD_CLEAR_MSG_FLAGS")
	case IPMI_CMD_GET_MSG_FLAGS:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_MSG_FLAGS")
	case IPMI_CMD_ENABLE_MESSAGE_CHANNEL_RCV:
		log.Println("      IPMI APP: Command = IPMI_CMD_ENABLE_MESSAGE_CHANNEL_RCV")
	case IPMI_CMD_GET_MSG:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_MSG")
	case IPMI_CMD_SEND_MSG:
		log.Println("      IPMI APP: Command = IPMI_CMD_SEND_MSG")
	case IPMI_CMD_READ_EVENT_MSG_BUFFER:
		log.Println("      IPMI APP: Command = IPMI_CMD_READ_EVENT_MSG_BUFFER")
	case IPMI_CMD_GET_BT_INTERFACE_CAPABILITIES:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_BT_INTERFACE_CAPABILITIES")
	case IPMI_CMD_GET_SYSTEM_GUID:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_SYSTEM_GUID")
	case IPMI_CMD_GET_CHANNEL_AUTH_CAPABILITIES:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_CHANNEL_AUTH_CAPABILITIES")
		HandleIPMIAuthenticationCapabilities(addr, server, wrapper, message)
	case IPMI_CMD_GET_SESSION_CHALLENGE:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_SESSION_CHALLENGE")
	case IPMI_CMD_ACTIVATE_SESSION:
		log.Println("      IPMI APP: Command = IPMI_CMD_ACTIVATE_SESSION")
	case IPMI_CMD_SET_SESSION_PRIVILEGE:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_SESSION_PRIVILEGE")
	case IPMI_CMD_CLOSE_SESSION:
		log.Println("      IPMI APP: Command = IPMI_CMD_CLOSE_SESSION")
	case IPMI_CMD_GET_SESSION_INFO:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_SESSION_INFO")
	case IPMI_CMD_GET_AUTHCODE:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_AUTHCODE")
	case IPMI_CMD_SET_CHANNEL_ACCESS:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_CHANNEL_ACCESS")
	case IPMI_CMD_GET_CHANNEL_ACCESS:
		log.Println("      IPMI APP: Command =IPMI_CMD_GET_CHANNEL_ACCESS")
	case IPMI_CMD_GET_CHANNEL_INFO:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_CHANNEL_INFO")
	case IPMI_CMD_SET_USER_ACCESS:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_USER_ACCESS")
	case IPMI_CMD_GET_USER_ACCESS:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_USER_ACCESS")
	case IPMI_CMD_SET_USER_NAME:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_USER_NAME")
	case IPMI_CMD_GET_USER_NAME:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_USER_NAME")
	case IPMI_CMD_SET_USER_PASSWORD:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_USER_PASSWORD")
	case IPMI_CMD_ACTIVATE_PAYLOAD:
		log.Println("      IPMI APP: Command = IPMI_CMD_ACTIVATE_PAYLOAD")
	case IPMI_CMD_DEACTIVATE_PAYLOAD:
		log.Println("      IPMI APP: Command = IPMI_CMD_DEACTIVATE_PAYLOAD")
	case IPMI_CMD_GET_PAYLOAD_ACTIVATION_STATUS:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_PAYLOAD_ACTIVATION_STATUS")
	case IPMI_CMD_GET_PAYLOAD_INSTANCE_INFO:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_PAYLOAD_INSTANCE_INFO")
	case IPMI_CMD_SET_USER_PAYLOAD_ACCESS:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_USER_PAYLOAD_ACCESS")
	case IPMI_CMD_GET_USER_PAYLOAD_ACCESS:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_USER_PAYLOAD_ACCESS")
	case IPMI_CMD_GET_CHANNEL_PAYLOAD_SUPPORT:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_CHANNEL_PAYLOAD_SUPPORT")
	case IPMI_CMD_GET_CHANNEL_PAYLOAD_VERSION:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_CHANNEL_PAYLOAD_VERSION")
	case IPMI_CMD_GET_CHANNEL_OEM_PAYLOAD_INFO:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_CHANNEL_OEM_PAYLOAD_INFO")
	case IPMI_CMD_MASTER_READ_WRITE:
		log.Println("      IPMI APP: Command = IPMI_CMD_MASTER_READ_WRITE")
	case IPMI_CMD_GET_CHANNEL_CIPHER_SUITES:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_CHANNEL_CIPHER_SUITES")
	case IPMI_CMD_SUSPEND_RESUME_PAYLOAD_ENCRYPTION:
		log.Println("      IPMI APP: Command = IPMI_CMD_SUSPEND_RESUME_PAYLOAD_ENCRYPTION")
	case IPMI_CMD_SET_CHANNEL_SECURITY_KEY:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_CHANNEL_SECURITY_KEY")
	case IPMI_CMD_GET_SYSTEM_INTERFACE_CAPABILITIES:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_SYSTEM_INTERFACE_CAPABILITIES")
	}
}