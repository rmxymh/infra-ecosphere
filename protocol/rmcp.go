package protocol

type RemoteManagementControlProtocol struct {
	Version uint8
	Reserved uint8
	Sequence uint8
	Class uint8
}

const (
	RMCP_VERSION_1	= 0x06
)

const (
	RMCP_CLASS_ASF	= 0x06
	RMCP_CLASS_IPMI	= 0x07
	RMCP_CLASS_OEM	= 0x08
)
