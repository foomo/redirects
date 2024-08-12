package redirectnats

import (
	"time"

	nc "github.com/nats-io/nats.go"
)

type NatsTopic string

const (
	ConfigNatsTimeout                 = 30 * time.Second
	ConfigNatsReconnectWait           = time.Second
	NatsTopicRedirects      NatsTopic = "redirects"
)

func DefaultNatsTopic() NatsTopic {
	return NatsTopicRedirects
}

func (t NatsTopic) String() string {
	return string(t)
}

func DefaultConnectOptions(clientID string) []nc.Option {
	return []nc.Option{
		nc.Name(clientID),
		nc.RetryOnFailedConnect(true),
		nc.Timeout(ConfigNatsTimeout),
		nc.ReconnectWait(ConfigNatsReconnectWait),
	}
}

func DefaultSubscribeOptions() []nc.SubOpt {
	return []nc.SubOpt{
		nc.DeliverAll(),
		nc.AckExplicit(),
	}
}
