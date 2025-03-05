package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/billbatista/ha-daikin-smart-ac-br/config"
	"github.com/billbatista/ha-daikin-smart-ac-br/daikin"
	"github.com/billbatista/ha-daikin-smart-ac-br/ha"
	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

func Server(ctx context.Context) error {
	config, err := config.NewConfig("./config.yaml")
	if err != nil {
		return err
	}

	mqttClient := pahomqtt.NewClient(
		pahomqtt.NewClientOptions().
			AddBroker(fmt.Sprintf("tcp://%s:%s", config.Mqtt.Host, config.Mqtt.Port)).
			SetUsername(config.Mqtt.Username).
			SetPassword(config.Mqtt.Password),
	)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		slog.Error("failed to connect", slog.Any("error", token.Error()))
		return token.Error()
	}
	slog.Info("connected to mqtt")

	defer func() {
		slog.Info("signal caught - exiting")
		mqttClient.Disconnect(1000)
		slog.Info("shutdown complete")
	}()

	for _, d := range config.Devices {
		url, err := url.Parse(d.Address)
		if err != nil {
			slog.Error("invalid target address", slog.Any("error", err), slog.String("address", d.Address))
			return err
		}
		secretKey, err := base64.StdEncoding.DecodeString(d.SecretKey)
		if err != nil {
			slog.Error("invalid secret key", slog.Any("error", err), slog.String("secretKey", d.SecretKey))
			return err
		}
		client := daikin.NewClient(url, secretKey)
		ac := ha.NewClimate(client, mqttClient, d.Name, d.UniqueId, d.OperationModes, d.FanModes)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			ac.PublishDiscovery()
			wg.Done()
		}()
		wg.Wait()

		_, err = client.State(ctx)
		if err != nil {
			slog.Error("could not get ac state", slog.String("name", d.UniqueId))
			ac.PublishUnavailable(ctx)
		}

		if err == nil {
			go func() {
				ac.PublishAvailable()
				ac.StateUpdate(ctx)
				ac.CommandSubscriptions()
			}()
		}
		defer func() {
			ac.PublishUnavailable(ctx)
		}()
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig

	return nil
}
