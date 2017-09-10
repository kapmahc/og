package nut

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
)

// JobHandler job handler
type JobHandler func(id, typ string, body []byte) error

var _jobHandlers = make(map[string]JobHandler)

// Receive receive jobs
func Receive(name string) error {
	log.Infof("waiting for messages, to exit press CTRL+C")
	return openAmqp(func(ch *amqp.Channel) error {
		if err := ch.Qos(1, 0, false); err != nil {
			return err
		}
		qu, err := ch.QueueDeclare(viper.GetString("server.name"), true, false, false, false, nil)
		if err != nil {
			return err
		}
		msgs, err := ch.Consume(qu.Name, name, false, false, false, false, nil)
		if err != nil {
			return err
		}
		for d := range msgs {
			d.Ack(false)
			now := time.Now()
			log.Infof("receive message %s@%s", d.MessageId, d.Type)
			hnd, ok := _jobHandlers[d.Type]
			if !ok {
				return fmt.Errorf("unknown message type %s", d.Type)
			}
			if err := hnd(d.MessageId, d.Type, d.Body); err != nil {
				return err
			}
			log.Infof("done %s %s", d.MessageId, time.Now().Sub(now))
		}
		return nil
	})
}

// Send send job
func Send(pri uint8, typ string, body []byte) error {
	return openAmqp(func(ch *amqp.Channel) error {
		qu, err := ch.QueueDeclare(viper.GetString("server.name"), true, false, false, false, nil)
		if err != nil {
			return err
		}

		return ch.Publish("", qu.Name, false, false, amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			MessageId:    uuid.New().String(),
			Priority:     pri,
			Body:         body,
			Timestamp:    time.Now(),
			Type:         typ,
		})
	})
}

func openAmqp(f func(*amqp.Channel) error) error {
	conn, err := amqp.Dial(
		fmt.Sprintf(
			"amqp://%s:%s@%s:%d/%s",
			viper.GetString("rabbitmq.user"),
			viper.GetString("rabbitmq.password"),
			viper.GetString("rabbitmq.host"),
			viper.GetInt("rabbitmq.port"),
			viper.GetString("rabbitmq.virtual"),
		),
	)
	if err != nil {
		return err
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return f(ch)
}
