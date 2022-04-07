package chanpubsub

import "fmt"

type Transport interface {
	// Receive gets a channel on which to receive messages with the specified name
	Receive(name string) <-chan []byte
	// Send gets a channel on which messages with the specified name may be sent.
	Send(name string) chan<- []byte
	// ErrChan gets a channel through which errors are sent
	ErrChan() <-chan error
	// Stop stops the transport. This will also close the channel returned by
	// Done method.
	Stop()
	// Done returns a channel which will close when calling the Stop() method.
	Done() chan struct{}
}

type TransportErr struct {
	Message []byte
	Name    string
	Err     error
}

func (te *TransportErr) Error() string {
	if len(te.Message) > 0 {
		return fmt.Sprintf("%s: |%s| <- `%s`", te.Err, te.Name, te.Message)
	}
	return fmt.Sprintf("%s: |%s|", te.Err, te.Name)
}
