package diplodocus

// Listener is a channel of events. It is safe
// to abandon channel that is not needed anymore.
type Listener chan event

// event has Buffer with data or an error.
// Channel should be abandoned after receiving an error.
type event struct {
	Buffer *[]byte
	Error  error
}
