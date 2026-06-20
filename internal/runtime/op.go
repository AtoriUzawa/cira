package runtime

type (
	// Op defines the operation type, used to identify WebSocket operations
	Op int
)

const (
	// OpRead represents a read operation.
	OpRead Op = iota // read message
	// OpWrite represents a write operation.
	OpWrite // write message

	// OpSendChan represents a send-to-channel operation.
	OpSendChan // send to channel

	// OpSetReadDeadline represents a read deadline setting operation.
	OpSetReadDeadline // set read deadline
	// OpSetWriteDeadline represents a write deadline setting operation.
	OpSetWriteDeadline // set write deadline

	// OpReadPong represents a pong frame read operation.
	OpReadPong // read Pong frame
	// OpWritePing represents a ping frame write operation.
	OpWritePing // write Ping frame

	// OpDecode represents a message decode operation.
	OpDecode // decode message
	// OpEncode represents a message encode operation.
	OpEncode // encode message
)

// String returns the human-readable string representation of Op
func (o Op) String() string {
	switch o {
	case OpRead:
		return "read"
	case OpWrite:
		return "write"
	case OpSendChan:
		return "send_chan"
	case OpSetReadDeadline:
		return "set_read_deadline"
	case OpSetWriteDeadline:
		return "set_write_deadline"
	case OpReadPong:
		return "read_pong"
	case OpWritePing:
		return "write_ping"
	case OpDecode:
		return "decode"
	case OpEncode:
		return "encode"
	default:
		return "unknown"
	}
}
