package nsq

import (
	"github.com/gagandeepahuja09/chanpubsub"
)

// make sure Transport satisfies vice.Transport interface.
var _ chanpubsub.Transport = (*Transport)(nil)

type Transport struct {
}
