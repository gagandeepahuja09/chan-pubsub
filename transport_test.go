package chanpubsub_test

import (
	"testing"
	"time"

	"github.com/gagandeepahuja09/chanpubsub"
)

// Test runs standard transport tests. All transport types pass this test.
// Transports should be passed with cleaned state. Old persisted messages can interfere
// with the tests.
// After the tests are run, the transport is closed.
func Transport(t *testing.T, transport func() chanpubsub.Transport) {
	testSendChannelsDontBlock(t, transport)
}

func testSendChannelsDontBlock(t *testing.T, newTransport func() chanpubsub.Transport) {
	transport := newTransport()

	select {
	case transport.Send("some_channel") <- []byte("message"):
	case <-time.After(1 * time.Second):
		t.Error("test send channel shouldn't block")
	}
}

// func testStandartTransportBehaviour(t *testing.T, transport func() chanpubsub.Transport) {
// 	transport := newTransport()
// 	transport1 := newTransport()
// 	transport2 := newTransport()
// }
