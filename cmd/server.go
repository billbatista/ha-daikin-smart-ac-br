package cmd

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/billbatista/ha-daikin-smart-ac-br/daikin"
	"github.com/billbatista/ha-daikin-smart-ac-br/ha"
	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	argSecretKey     string
	argTargetAddress string
)

func Server(ctx context.Context) error {
	flag.StringVar(&argSecretKey, "secretKey", "", "The secret key")
	flag.StringVar(&argTargetAddress, "targetAddress", "", "The target address")
	flag.Parse()

	url, err := url.Parse(argTargetAddress)
	if err != nil {
		fmt.Printf("invalid target address: %v\n", err)
		return err
	}
	secretKey, err := base64.StdEncoding.DecodeString(argSecretKey)
	if err != nil {
		fmt.Printf("could not decode the secret key: %v\n", err)
		return err
	}

	client := daikin.NewClient(url, secretKey)

	mqttClient := pahomqtt.NewClient(
		pahomqtt.NewClientOptions().
			AddBroker("tcp://10.1.1.6:1883").
			SetUsername("mqtt").
			SetPassword("mqtt_password_2023"),
	)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		slog.Error("failed to connect", slog.Any("error", token.Error()))
		return token.Error()
	}
	defer func() {
		fmt.Println("signal caught - exiting")
		mqttClient.Disconnect(1000)
		fmt.Println("shutdown complete")
	}()

	_, err = client.State(ctx)
	if err != nil {
		fmt.Print("could not get ac state")
		return err
	}

	ac := ha.NewClimate(client, mqttClient, "Daikin Suíte", "DAIKIN46ACB4", ha.DefaultOperationModes, ha.DefaultFanModes)
	ac.PublishDiscovery()
	ac.StateUpdate(ctx)
	ac.CommandSubscriptions()

	// Grab info from config (yaml) or from aws/iotalabs things
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig

	return nil
}
