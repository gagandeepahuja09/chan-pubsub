package transporttest

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gagandeepahuja09/chanpubsub"
	"github.com/stretchr/testify/assert"
)

// Test runs standard transport tests. All transport types pass this test.
// Transports should be passed with cleaned state. Old persisted messages can interfere
// with the tests.
// After the tests are run, the transport is closed.
func Transport(t *testing.T, transport func() chanpubsub.Transport) {
	testSendChannelsDontBlock(t, transport)
	testStandardTransportBehaviour(t, transport)
}

func testSendChannelsDontBlock(t *testing.T, newTransport func() chanpubsub.Transport) {
	transport := newTransport()

	select {
	case transport.Send("some_channel") <- []byte("message"):
	case <-time.After(1 * time.Second):
		t.Error("test send channel shouldn't block")
	}
}

func testStandardTransportBehaviour(t *testing.T, newTransport func() chanpubsub.Transport) {
	transport := newTransport()
	transport1 := newTransport()
	transport2 := newTransport()

	doneChan := make(chan struct{})
	messages := make(map[string][][]byte)
	var wg sync.WaitGroup

	go func() {
		defer close(doneChan)
		for {
			select {
			case <-transport.Done():
				return
			case err := <-transport.ErrChan():
				assert.NoError(t, err)

			// test local load balancing with the same transport
			case msg := <-transport.Receive("vicechannel1"):
				messages["vicechannel1"] = append(messages["vicechannel1"], msg)
				wg.Done()

			case msg := <-transport.Receive("vicechannel2"):
				messages["vicechannel2"] = append(messages["vicechannel2"], msg)
				wg.Done()
			case msg := <-transport.Receive("vicechannel2"):
				messages["vicechannel2"] = append(messages["vicechannel2"], msg)
				wg.Done()

			case msg := <-transport.Receive("vicechannel3"):
				messages["vicechannel3"] = append(messages["vicechannel3"], msg)
				wg.Done()
			case msg := <-transport.Receive("vicechannel3"):
				messages["vicechannel3"] = append(messages["vicechannel3"], msg)
				wg.Done()
			case msg := <-transport.Receive("vicechannel3"):
				messages["vicechannel3"] = append(messages["vicechannel3"], msg)
				wg.Done()
			case msg := <-transport.Receive("vicechannel3"):
				messages["vicechannel3"] = append(messages["vicechannel3"], msg)
				wg.Done()

			// test distributed load balancing
			case msg := <-transport.Receive("vicechannel4"):
				messages["vicechannel4.1"] = append(messages["vicechannel4.1"], msg)
				wg.Done()
			case msg := <-transport1.Receive("vicechannel3"):
				messages["vicechannel4.2"] = append(messages["vicechannel4.2"], msg)
				wg.Done()
			case msg := <-transport2.Receive("vicechannel3"):
				messages["vicechannel4.3"] = append(messages["vicechannel4.3"], msg)
				wg.Done()
			}
		}
	}()

	// Time to initialize receiving channels
	time.Sleep(time.Millisecond * 10)

	// send 100 messages down each channel
	for i := 0; i < 100; i++ {
		wg.Add(4)
		msg := []byte(fmt.Sprintf("message %d", i+1))
		transport.Send("vicechannel1") <- msg
		transport.Send("vicechannel2") <- msg
		transport.Send("vicechannel3") <- msg
		transport.Send("vicechannel4") <- msg
	}

	wg.Wait()
	transport.Stop()
	transport1.Stop()
	transport2.Stop()
	<-doneChan

	assert.Equal(t, 6, len(messages))
	assert.Equal(t, 100, len(messages["vicechannel1"]))
	assert.Equal(t, 100, len(messages["vicechannel2"]))
	assert.Equal(t, 100, len(messages["vicechannel3"]))

	assert.NotEqual(t, 100, len(messages["vicechannel4.1"]))
	assert.NotEqual(t, 100, len(messages["vicechannel4.2"]))
	assert.NotEqual(t, 100, len(messages["vicechannel4.3"]))

	assert.Equal(t, 100, len(messages["vicechannel4.1"])+len(messages["vicechannel4.2"])+len(messages["vicechannel4.3"]))
}
