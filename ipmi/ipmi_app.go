package ipmi

import (
	"net"
	"bytes"
	"log"
	"encoding/binary"
	"fmt"
	"math/rand"
	"crypto/md5"
)

import (
	"github.com/htruong/go-md2"
	"github.com/rmxymh/infra-ecosphere/bmc"
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
	AUTH_NONE =	0x00
	AUTH_MD2 =	0x01
	AUTH_MD5 =	0x02
)

const (
	AUTH_BITMASK_NONE = 		0x01
	AUTH_BITMASK_MD2 = 		0x02
	AUTH_BITMASK_MD5 = 		0x04
	AUTH_BITMASK_STRAIGHT_KEY =	0x10
	AUTH_BITMASK_OEM = 		0x20
	AUTH_BITMASK_IPMI_V2 = 		0x80
)

const (
	AUTH_STATUS_ANONYMOUS =		0x01
	AUTH_STATUS_NULL_USER =		0x02
	AUTH_STATUS_NON_NULL_USER =	0x04
	AUTH_STATUS_USER_LEVEL = 	0x08
	AUTH_STATUS_PER_MESSAGE = 	0x10
	AUTH_STATUS_KG =		0x20
)

const (
	COMPLETION_CODE_OK = 			0x00
	COMPLETION_CODE_INVALID_USERNAME =	0x81
)

func dumpByteBuffer(buf bytes.Buffer) {
	bytebuf := buf.Bytes()
	fmt.Print("[")
	for i := 0 ; i < buf.Len(); i += 1 {
		fmt.Printf(" %02x", bytebuf[i])
	}
	fmt.Println("]")
}

func BuildResponseMessageTemplate(requestWrapper IPMISessionWrapper, requestMessage IPMIMessage,  netfn uint8, command uint8) (IPMISessionWrapper, IPMIMessage) {
	responseMessage := IPMIMessage{}
	responseMessage.TargetAddress = requestMessage.SourceAddress
	remoteLun := requestMessage.SourceLun & 0x03
	localLun := requestMessage.TargetLun & 0x03
	responseMessage.TargetLun = remoteLun | (netfn << 2)
	responseMessage.SourceAddress = requestMessage.TargetAddress
	responseMessage.SourceLun = (requestMessage.SourceLun & 0xfc) | localLun
	responseMessage.Command = command
	responseMessage.CompletionCode = COMPLETION_CODE_OK

	responseWrapper := IPMISessionWrapper{}
	responseWrapper.AuthenticationType = requestWrapper.AuthenticationType
	responseWrapper.SequenceNumber = 0xff
	responseWrapper.SessionId = requestWrapper.SessionId

	return responseWrapper, responseMessage
}

func HandleIPMIUnsupportedAppCommand(addr *net.UDPAddr, server *net.UDPConn, wrapper IPMISessionWrapper, message IPMIMessage) {
	log.Println("      IPMI App: This command is not supported currently, ignore.")
}

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

func HandleIPMIAuthenticationCapabilities(addr *net.UDPAddr, server *net.UDPConn, wrapper IPMISessionWrapper, message IPMIMessage) {
	buf := bytes.NewBuffer(message.Data)
	request := IPMIAuthenticationCapabilitiesRequest{}
	binary.Read(buf, binary.BigEndian, &request)

	// prepare for response data
	// We don't simulate OEM related behavior
	response := IPMIAuthenticationCapabilitiesResponse{}
	response.Channel = 1
	response.AuthenticationTypeSupport = AUTH_BITMASK_MD5 | AUTH_BITMASK_MD2 | AUTH_BITMASK_NONE
	response.AuthenticationStatus = AUTH_STATUS_NON_NULL_USER | AUTH_STATUS_NULL_USER
	response.ExtCapabilities = 0
	response.OEMAuxiliaryData = 0

	dataBuf := bytes.Buffer{}
	binary.Write(&dataBuf, binary.BigEndian, response)

	responseWrapper, responseMessage := BuildResponseMessageTemplate(wrapper, message, (IPMI_NETFN_APP | IPMI_NETFN_RESPONSE), IPMI_CMD_GET_CHANNEL_AUTH_CAPABILITIES)
	responseMessage.Data = dataBuf.Bytes()
	rmcp := BuildUpRMCPForIPMI()

	// serialize and send back
	obuf := bytes.Buffer{}
	SerializeRMCP(&obuf, rmcp)
	SerializeIPMI(&obuf, responseWrapper, responseMessage)

	server.WriteToUDP(obuf.Bytes(), addr)
}

type IPMIGetSessionChallengeRequest struct {
	AuthenticationType uint8
	Username [16]byte
}

type IPMIGetSessionChallengeResponse struct {
	TempSessionID uint32
	Challenge [16]byte
}

func HandleIPMIGetSessionChallenge(addr *net.UDPAddr, server *net.UDPConn, wrapper IPMISessionWrapper, message IPMIMessage) {
	buf := bytes.NewBuffer(message.Data)
	request := IPMIGetSessionChallengeRequest{}
	binary.Read(buf, binary.BigEndian, &request)

	obuf := bytes.Buffer{}

	nameLength := len(request.Username)
	for i := range request.Username {
		if request.Username[i] == 0 {
			nameLength = i
			break
		}
	}
	username := string(request.Username[:nameLength])

	user, found := bmc.GetBMCUser(username)
	if ! found {
		responseWrapper, responseMessage := BuildResponseMessageTemplate(wrapper, message, (IPMI_NETFN_APP | IPMI_NETFN_RESPONSE), IPMI_CMD_GET_SESSION_CHALLENGE)
		responseMessage.CompletionCode = COMPLETION_CODE_INVALID_USERNAME
		rmcp := BuildUpRMCPForIPMI()

		SerializeRMCP(&obuf, rmcp)
		SerializeIPMI(&obuf, responseWrapper, responseMessage)
	} else {
		session := GetNewSession(user)
		var challengeCode [16]uint8

		for i := range challengeCode {
			challengeCode[i] = uint8(rand.Uint32() % 0xff)
		}

		responseChallenge := IPMIGetSessionChallengeResponse{}
		responseChallenge.TempSessionID = session.SessionID
		responseChallenge.Challenge = challengeCode
		dataBuf := bytes.Buffer{}
		binary.Write(&dataBuf, binary.BigEndian, responseChallenge)

		responseWrapper, responseMessage := BuildResponseMessageTemplate(wrapper, message, (IPMI_NETFN_APP | IPMI_NETFN_RESPONSE), IPMI_CMD_GET_SESSION_CHALLENGE)
		responseMessage.Data = dataBuf.Bytes()
		rmcp := BuildUpRMCPForIPMI()

		SerializeRMCP(&obuf, rmcp)
		SerializeIPMI(&obuf, responseWrapper, responseMessage)
	}

	server.WriteToUDP(obuf.Bytes(), addr)
}

type IPMIActivateSessionRequest struct {
	AuthenticationType uint8
	RequestMaxPrivilegeLevel uint8
	Challenge [16]byte
	InitialOutboundSeq uint32
}

type IPMIActivateSessionResponse struct {
	AuthenticationType uint8
	SessionId uint32
	InitialOutboundSeq uint32
	MaxPrivilegeLevel uint8
}

func GetAuthenticationCode(authenticationType uint8, password string, sessionID uint32, message IPMIMessage, sessionSeq uint32) [16]byte {
	var passwordBytes [16]byte
	copy(passwordBytes[:], password)

	context := bytes.Buffer{}
	binary.Write(&context, binary.BigEndian, passwordBytes)
	binary.Write(&context, binary.BigEndian, sessionID)
	SerializeIPMIMessage(&context, message)
	binary.Write(&context, binary.BigEndian, sessionSeq)
	binary.Write(&context, binary.BigEndian, passwordBytes)

	var code [16]byte
	switch authenticationType {
	case AUTH_MD5:
		code = md5.Sum(context.Bytes())
	case AUTH_MD2:
		hash := md2.New()
		md2Code := hash.Sum(context.Bytes())
		for i := range md2Code {
			if i >= len(code) {
				break
			}
			code[i] = md2Code[i]
		}
	}

	return code
}

func HandleIPMIActivateSession(addr *net.UDPAddr, server *net.UDPConn, wrapper IPMISessionWrapper, message IPMIMessage) {
	buf := bytes.NewBuffer(message.Data)
	request := IPMIActivateSessionRequest{}
	binary.Read(buf, binary.BigEndian, &request)

	//obuf := bytes.Buffer{}

	session, ok := GetSession(wrapper.SessionId)
	if ! ok {
		log.Printf("Unable to find session 0x%08x\n", wrapper.SessionId)
	} else {
		bmcUser := session.User
		code := GetAuthenticationCode(wrapper.AuthenticationType, bmcUser.Password, wrapper.SessionId, message, wrapper.SequenceNumber)
		if bytes.Compare(wrapper.AuthenticationCode[:], code[:]) == 0 {
			log.Println("      IPMI Authentication Pass.")
		} else {
			log.Println("      IPMI Authentication Failed.")
		}

		session.RemoteSessionSequenceNumber = request.InitialOutboundSeq
		session.LocalSessionSequenceNumber = 0

		response := IPMIActivateSessionResponse{}
		response.AuthenticationType = request.AuthenticationType
		response.SessionId = wrapper.SessionId
		session.LocalSessionSequenceNumber += 1
		response.InitialOutboundSeq = session.LocalSessionSequenceNumber
		response.MaxPrivilegeLevel = request.RequestMaxPrivilegeLevel

		dataBuf := bytes.Buffer{}
		binary.Write(&dataBuf, binary.BigEndian, response)

		responseWrapper, responseMessage := BuildResponseMessageTemplate(wrapper, message, (IPMI_NETFN_APP | IPMI_NETFN_RESPONSE), IPMI_CMD_ACTIVATE_SESSION)
		responseMessage.Data = dataBuf.Bytes()

		responseWrapper.SessionId = response.SessionId
		responseWrapper.SequenceNumber = session.RemoteSessionSequenceNumber
		responseWrapper.AuthenticationCode = GetAuthenticationCode(response.AuthenticationType, bmcUser.Password, response.SessionId, responseMessage, responseWrapper.SequenceNumber)
		rmcp := BuildUpRMCPForIPMI()

		obuf := bytes.Buffer{}
		SerializeRMCP(&obuf, rmcp)
		SerializeIPMI(&obuf, responseWrapper, responseMessage)
		server.WriteToUDP(obuf.Bytes(), addr)
	}
}

type IPMISetSessionPrivilegeLevelRequest struct {
	RequestPrivilegeLevel uint8
}

type IPMISetSessionPrivilegeLevelResponse struct {
	NewPrivilegeLevel uint8
}

func HandleIPMISetSessionPrivilegeLevel(addr *net.UDPAddr, server *net.UDPConn, wrapper IPMISessionWrapper, message IPMIMessage) {
	buf := bytes.NewBuffer(message.Data)
	request := IPMISetSessionPrivilegeLevelRequest{}
	binary.Read(buf, binary.BigEndian, &request)

	//obuf := bytes.Buffer{}

	session, ok := GetSession(wrapper.SessionId)
	if ! ok {
		log.Printf("Unable to find session 0x%08x\n", wrapper.SessionId)
	} else {
		bmcUser := session.User
		code := GetAuthenticationCode(wrapper.AuthenticationType, bmcUser.Password, wrapper.SessionId, message, wrapper.SequenceNumber)
		if bytes.Compare(wrapper.AuthenticationCode[:], code[:]) == 0 {
			log.Println("      IPMI Authentication Pass.")
		} else {
			log.Println("      IPMI Authentication Failed.")
		}

		session.LocalSessionSequenceNumber += 1
		session.RemoteSessionSequenceNumber += 1

		response := IPMISetSessionPrivilegeLevelResponse{}
		response.NewPrivilegeLevel = request.RequestPrivilegeLevel

		dataBuf := bytes.Buffer{}
		binary.Write(&dataBuf, binary.BigEndian, response)

		responseWrapper, responseMessage := BuildResponseMessageTemplate(wrapper, message, (IPMI_NETFN_APP | IPMI_NETFN_RESPONSE), IPMI_CMD_SET_SESSION_PRIVILEGE)
		responseMessage.Data = dataBuf.Bytes()

		responseWrapper.SessionId = wrapper.SessionId
		responseWrapper.SequenceNumber = session.RemoteSessionSequenceNumber
		responseWrapper.AuthenticationCode = GetAuthenticationCode(wrapper.AuthenticationType, bmcUser.Password, responseWrapper.SessionId, responseMessage, responseWrapper.SequenceNumber)
		rmcp := BuildUpRMCPForIPMI()

		obuf := bytes.Buffer{}
		SerializeRMCP(&obuf, rmcp)
		SerializeIPMI(&obuf, responseWrapper, responseMessage)
		server.WriteToUDP(obuf.Bytes(), addr)
	}
}

type IPMICloseSessionRequest struct {
	SessionID uint32
}

func HandleIPMICloseSession(addr *net.UDPAddr, server *net.UDPConn, wrapper IPMISessionWrapper, message IPMIMessage) {
	buf := bytes.NewBuffer(message.Data)
	request := IPMICloseSessionRequest{}
	binary.Read(buf, binary.BigEndian, &request)

	//obuf := bytes.Buffer{}

	session, ok := GetSession(wrapper.SessionId)
	if ! ok {
		log.Printf("Unable to find session 0x%08x\n", wrapper.SessionId)
	} else {
		bmcUser := session.User
		code := GetAuthenticationCode(wrapper.AuthenticationType, bmcUser.Password, wrapper.SessionId, message, wrapper.SequenceNumber)
		if bytes.Compare(wrapper.AuthenticationCode[:], code[:]) == 0 {
			log.Println("      IPMI Authentication Pass.")
		} else {
			log.Println("      IPMI Authentication Failed.")
		}

		RemoveSession(request.SessionID)

		session.LocalSessionSequenceNumber += 1
		session.RemoteSessionSequenceNumber += 1

		responseWrapper, responseMessage := BuildResponseMessageTemplate(wrapper, message, (IPMI_NETFN_APP | IPMI_NETFN_RESPONSE), IPMI_CMD_CLOSE_SESSION)

		responseWrapper.SessionId = wrapper.SessionId
		responseWrapper.SequenceNumber = session.RemoteSessionSequenceNumber
		responseWrapper.AuthenticationCode = GetAuthenticationCode(wrapper.AuthenticationType, bmcUser.Password, responseWrapper.SessionId, responseMessage, responseWrapper.SequenceNumber)
		rmcp := BuildUpRMCPForIPMI()

		obuf := bytes.Buffer{}
		SerializeRMCP(&obuf, rmcp)
		SerializeIPMI(&obuf, responseWrapper, responseMessage)
		server.WriteToUDP(obuf.Bytes(), addr)
	}
}

func IPMI_APP_DeserializeAndExecute(addr *net.UDPAddr, server *net.UDPConn, wrapper IPMISessionWrapper, message IPMIMessage) {
	switch message.Command {
	case IPMI_CMD_GET_DEVICE_ID:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_DEVICE_ID")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_COLD_RESET:
		log.Println("      IPMI APP: Command = IPMI_CMD_COLD_RESET")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_WARM_RESET:
		log.Println("      IPMI APP: Command = IPMI_CMD_WARM_RESET")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_SELF_TEST_RESULTS:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_SELF_TEST_RESULTS")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_MANUFACTURING_TEST_ON:
		log.Println("      IPMI APP: Command = IPMI_CMD_MANUFACTURING_TEST_ON")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_SET_ACPI_POWER_STATE:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_ACPI_POWER_STATE")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_ACPI_POWER_STATE:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_ACPI_POWER_STATE")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_DEVICE_GUID:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_DEVICE_GUID")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_RESET_WATCHDOG_TIMER:
		log.Println("      IPMI APP: Command = IPMI_CMD_RESET_WATCHDOG_TIMER")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_SET_WATCHDOG_TIMER:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_WATCHDOG_TIMER")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_WATCHDOG_TIMER:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_WATCHDOG_TIMER")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_SET_BMC_GLOBAL_ENABLES:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_BMC_GLOBAL_ENABLES")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_BMC_GLOBAL_ENABLES:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_BMC_GLOBAL_ENABLES")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_CLEAR_MSG_FLAGS:
		log.Println("      IPMI APP: Command =IPMI_CMD_CLEAR_MSG_FLAGS")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_MSG_FLAGS:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_MSG_FLAGS")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_ENABLE_MESSAGE_CHANNEL_RCV:
		log.Println("      IPMI APP: Command = IPMI_CMD_ENABLE_MESSAGE_CHANNEL_RCV")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_MSG:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_MSG")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_SEND_MSG:
		log.Println("      IPMI APP: Command = IPMI_CMD_SEND_MSG")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_READ_EVENT_MSG_BUFFER:
		log.Println("      IPMI APP: Command = IPMI_CMD_READ_EVENT_MSG_BUFFER")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_BT_INTERFACE_CAPABILITIES:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_BT_INTERFACE_CAPABILITIES")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_SYSTEM_GUID:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_SYSTEM_GUID")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_CHANNEL_AUTH_CAPABILITIES:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_CHANNEL_AUTH_CAPABILITIES")
		HandleIPMIAuthenticationCapabilities(addr, server, wrapper, message)
	case IPMI_CMD_GET_SESSION_CHALLENGE:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_SESSION_CHALLENGE")
		HandleIPMIGetSessionChallenge(addr, server, wrapper, message)
	case IPMI_CMD_ACTIVATE_SESSION:
		log.Println("      IPMI APP: Command = IPMI_CMD_ACTIVATE_SESSION")
		HandleIPMIActivateSession(addr, server, wrapper, message)
	case IPMI_CMD_SET_SESSION_PRIVILEGE:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_SESSION_PRIVILEGE")
		HandleIPMISetSessionPrivilegeLevel(addr, server, wrapper, message)
	case IPMI_CMD_CLOSE_SESSION:
		log.Println("      IPMI APP: Command = IPMI_CMD_CLOSE_SESSION")
		HandleIPMICloseSession(addr, server, wrapper, message)
	case IPMI_CMD_GET_SESSION_INFO:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_SESSION_INFO")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_AUTHCODE:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_AUTHCODE")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_SET_CHANNEL_ACCESS:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_CHANNEL_ACCESS")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_CHANNEL_ACCESS:
		log.Println("      IPMI APP: Command =IPMI_CMD_GET_CHANNEL_ACCESS")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_CHANNEL_INFO:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_CHANNEL_INFO")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_SET_USER_ACCESS:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_USER_ACCESS")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_USER_ACCESS:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_USER_ACCESS")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_SET_USER_NAME:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_USER_NAME")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_USER_NAME:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_USER_NAME")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_SET_USER_PASSWORD:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_USER_PASSWORD")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_ACTIVATE_PAYLOAD:
		log.Println("      IPMI APP: Command = IPMI_CMD_ACTIVATE_PAYLOAD")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_DEACTIVATE_PAYLOAD:
		log.Println("      IPMI APP: Command = IPMI_CMD_DEACTIVATE_PAYLOAD")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_PAYLOAD_ACTIVATION_STATUS:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_PAYLOAD_ACTIVATION_STATUS")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_PAYLOAD_INSTANCE_INFO:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_PAYLOAD_INSTANCE_INFO")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_SET_USER_PAYLOAD_ACCESS:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_USER_PAYLOAD_ACCESS")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_USER_PAYLOAD_ACCESS:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_USER_PAYLOAD_ACCESS")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_CHANNEL_PAYLOAD_SUPPORT:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_CHANNEL_PAYLOAD_SUPPORT")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_CHANNEL_PAYLOAD_VERSION:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_CHANNEL_PAYLOAD_VERSION")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_CHANNEL_OEM_PAYLOAD_INFO:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_CHANNEL_OEM_PAYLOAD_INFO")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_MASTER_READ_WRITE:
		log.Println("      IPMI APP: Command = IPMI_CMD_MASTER_READ_WRITE")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_CHANNEL_CIPHER_SUITES:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_CHANNEL_CIPHER_SUITES")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_SUSPEND_RESUME_PAYLOAD_ENCRYPTION:
		log.Println("      IPMI APP: Command = IPMI_CMD_SUSPEND_RESUME_PAYLOAD_ENCRYPTION")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_SET_CHANNEL_SECURITY_KEY:
		log.Println("      IPMI APP: Command = IPMI_CMD_SET_CHANNEL_SECURITY_KEY")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	case IPMI_CMD_GET_SYSTEM_INTERFACE_CAPABILITIES:
		log.Println("      IPMI APP: Command = IPMI_CMD_GET_SYSTEM_INTERFACE_CAPABILITIES")
		HandleIPMIUnsupportedAppCommand(addr, server, wrapper, message)

	}
}