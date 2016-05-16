package protocol

type AlertStandardFormat struct {
	IANA uint32
	MessageType uint8
	MessageTag uint8
	Reserved uint8
	DataLen uint8
}

const (
	ASF_RMCP_IANA	= 0x000011be
)

const (
	ASF_TYPE_PING	= 0x80
	ASF_TYPE_PONG	= 0x40
)
