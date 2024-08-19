package redirectnats

import (
	"context"
	"encoding/json"

	keellog "github.com/foomo/keel/log"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type UpdateSignal struct {
	topic        string
	l            *zap.Logger
	connection   *nats.Conn
	subscription *nats.Subscription
	messages     chan *nats.Msg
}

func NewUpdateSignalSubscribeChannel(
	ctx context.Context,
	l *zap.Logger,
	natsURI,
	clientID,
	topic string,
) (chan *nats.Msg, error) {
	updateSignal, err := NewUpdateSignal(ctx, l, natsURI, clientID, topic)
	if err != nil {
		return nil, err
	}

	channel, err := updateSignal.Subscribe()
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func NewUpdateSignal(ctx context.Context, l *zap.Logger, natsURI, clientID, topic string) (*UpdateSignal, error) {
	var err error
	c := &UpdateSignal{
		topic:    topic,
		l:        l,
		messages: make(chan *nats.Msg),
	}
	c.connection, err = nats.Connect(natsURI, DefaultConnectOptions(clientID)...)
	if err != nil {
		keellog.WithError(c.l, err).Error("error when connecting to nats")
		return nil, err
	}
	go func() {
		<-ctx.Done()
		_ = c.Close(ctx)
	}()
	return c, nil
}

func (c *UpdateSignal) Close(_ context.Context) error {
	if c.connection != nil {
		err := c.subscription.Unsubscribe()
		if err != nil {
			keellog.WithError(c.l, err).Error("error when unsubscribing")
			return err
		}
		close(c.messages)
		c.connection.Close()
	}
	return nil
}

func (c *UpdateSignal) Subscribe() (chan *nats.Msg, error) {
	subscription, err := c.connection.ChanSubscribe(c.topic, c.messages)
	if err != nil {
		keellog.WithError(c.l, err).Error("error when subscribing")
		return nil, err
	}
	c.subscription = subscription
	return c.messages, nil
}

func (c *UpdateSignal) Publish() error {
	payload, _ := json.Marshal(struct{}{})
	return c.connection.Publish(c.topic, payload)
}
