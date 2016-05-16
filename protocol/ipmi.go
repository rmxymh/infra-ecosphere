package protocol

// port from OpenIPMI
// Network Functions
const (
	IPMI_NETFN_CHASSIS =		0x00
	IPMI_NETFN_BRIDGE =		0x02
	IPMI_NETFN_SENSOR_EVENT =	0x04
	IPMI_NETFN_APP =			0x06
	IPMI_NETFN_FIRMWARE =		0x08
	IPMI_NETFN_STORAGE	 =		0x0a
	IPMI_NETFN_TRANSPORT =		0x0c
	IPMI_NETFN_GROUP_EXTENSION =	0x2c
	IPMI_NETFN_OEM_GROUP =		0x2e
)

// Authentication Type
const (
	AUTH_NONE =		0x00
	AUTH_MD2 =		0x01
	AUTH_MD5 =		0x02
	AUTH_STRAIGHT = 	0x04
	AUTH_OEM = 		0x05
	AUTH_RMCP_PLUS =	0x06		// IPMI v2.0
)

type IPMISessionWrapper struct {
	AuthenticationType uint8
	SequenceNumber uint32
	SessionId uint32
	MessageLen uint8
}

type IPMIMessage struct {
	TargetAddress uint8
	TargetLun uint8			// NetFn (6) + Lun (2)
	Checksum uint8
	SourceAddress uint8
	SourceLun uint8			// SequenceNumber (6) + Lun (2)
	Command uint8
}

type IPMIAuthenticationCapabilitiesRequest struct {
	VersionCompatibliity uint8
	RequestedPrivilegeLevel uint8
	DataChecksum uint8
}

