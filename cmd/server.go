package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
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

	url, err := url.Parse(config.Devices[0].Address)
	if err != nil {
		fmt.Printf("invalid target address: %v\n", err)
		return err
	}
	secretKey, err := base64.StdEncoding.DecodeString(config.Devices[0].SecretKey)
	if err != nil {
		fmt.Printf("could not decode the secret key: %v\n", err)
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
	defer func() {
		slog.Info("signal caught - exiting")
		mqttClient.Disconnect(1000)
		slog.Info("shutdown complete")
	}()

	for _, d := range config.Devices {
		client := daikin.NewClient(url, secretKey)
		_, err = client.State(ctx)
		if err != nil {
			fmt.Print("could not get ac state")
			return err
		}

		ac := ha.NewClimate(client, mqttClient, d.Name, d.UniqueId, d.OperationModes, d.FanModes)
		ac.PublishDiscovery()
		ac.StateUpdate(ctx)
		ac.CommandSubscriptions()

		defer func() {
			ac.PublishUnavailable(ctx)
		}()
	}

	// Grab info from config (yaml) or from aws/iotalabs things
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig

	return nil
}
