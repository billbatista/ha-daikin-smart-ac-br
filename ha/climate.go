package ha

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/billbatista/ha-daikin-smart-ac-br/daikin"
	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	DefaultFanModes       = []string{"auto", "low", "medium", "high"}
	DefaultOperationModes = []string{"auto", "off", "cool", "heat", "dry", "fan_only"}
)

type Climate struct {
	Name                         string   `json:"name"`
	UniqueId                     string   `json:"unique_id"`
	Modes                        []string `json:"modes"`
	ModeCommandTopic             string   `json:"mode_command_topic"`
	ModeStateTopic               string   `json:"mode_state_topic"`
	FanModes                     []string `json:"fan_modes"`
	FanModeCommandTopic          string   `json:"fan_mode_command_topic"`
	FanModeStateTopic            string   `json:"fan_mode_state_topic"`
	TemperatureCommandTopic      string   `json:"temperature_command_topic"`
	TemperatureStateTopic        string   `json:"temperature_state_topic"`
	CurrentTemperatureStateTopic string   `json:"current_temperature_topic"`
	TemperatureUnit              string   `json:"temperature_unit"`
	Precision                    float32  `json:"precision"`
	SwingModeStateTopic          string   `json:"swing_mode_state_topic"`
	SwingModeCommandTopic        string   `json:"swing_mode_command_topic"`
	SwingModes                   []string `json:"swing_modes"`
	Device                       Device   `json:"device"`
	daikinClient                 *daikin.Client
	mqtt                         pahomqtt.Client
}

type Device struct {
	Name         string `json:"name"`
	Ids          string `json:"ids"`
	Manufacturer string `json:"manufacturer"`
}

func NewClimate(daikinClient *daikin.Client, mqttClient pahomqtt.Client, name string, uniqueId string, modes []string, fanModes []string) *Climate {
	if len(modes) == 0 {
		modes = DefaultOperationModes
	}
	if len(fanModes) == 0 {
		fanModes = DefaultFanModes
	}
	uniqueId = strings.ToLower(uniqueId)

	return &Climate{
		daikinClient:                 daikinClient,
		mqtt:                         mqttClient,
		Name:                         "Ar Condicionado",
		UniqueId:                     uniqueId,
		Modes:                        modes,
		ModeCommandTopic:             fmt.Sprintf("daikin/%s/mode/set", uniqueId),
		ModeStateTopic:               fmt.Sprintf("daikin/%s/mode/state", uniqueId),
		FanModes:                     fanModes,
		FanModeCommandTopic:          fmt.Sprintf("daikin/%s/fan_mode/set", uniqueId),
		FanModeStateTopic:            fmt.Sprintf("daikin/%s/fan_mode/state", uniqueId),
		TemperatureUnit:              "C",
		Precision:                    float32(1),
		TemperatureCommandTopic:      fmt.Sprintf("daikin/%s/target_temperature/set", uniqueId),
		TemperatureStateTopic:        fmt.Sprintf("daikin/%s/target_temperature/state", uniqueId),
		CurrentTemperatureStateTopic: fmt.Sprintf("daikin/%s/temperature/state", uniqueId),
		SwingModes:                   []string{"on", "off"},
		SwingModeCommandTopic:        fmt.Sprintf("daikin/%s/swing_mode/set", uniqueId),
		SwingModeStateTopic:          fmt.Sprintf("daikin/%s/swing_mode/state", uniqueId),
		Device: Device{
			Name:         name,
			Ids:          uniqueId,
			Manufacturer: "Daikin Brazil",
		},
	}
}

func (c *Climate) StateUpdate(ctx context.Context) {
	stateCh := make(chan *daikin.State)
	go func() {
		for {
			start := time.Now()
			state, err := c.daikinClient.State(ctx)
			duration := time.Since(start)
			if err != nil {
				slog.ErrorContext(ctx, "failed to get ac state", slog.String("name", c.Device.Name))
			}
			slog.InfoContext(ctx, "retrieved ac state", slog.String("name", c.Device.Name), slog.Any("duration", duration))
			stateCh <- state
			time.Sleep(1 * time.Second)
		}
	}()
	go func() {
		for {
			v, ok := <-stateCh
			if !ok {
				slog.InfoContext(ctx, "channel closed")
				return
			}
			fanMode := c.parseFanMode(v.Port1.Fan)
			token := c.mqtt.Publish(c.FanModeStateTopic, 0, false, fanMode)
			if token.Error() != nil {
				slog.ErrorContext(ctx, "failed to publish ac fan mode state", slog.Any("error", token.Error()))
			}
			slog.InfoContext(ctx, "fan mode updated", slog.String("fan_mode", fanMode), slog.String("name", c.Device.Name))

			currentTemp := strconv.FormatFloat(v.Port1.Sensors.RoomTemp, 'f', -1, 64)
			token = c.mqtt.Publish(c.CurrentTemperatureStateTopic, 0, false, currentTemp)
			if token.Error() != nil {
				slog.ErrorContext(ctx, "failed to publish ac current temperature state", slog.Any("error", token.Error()))
			}
			slog.InfoContext(ctx, "current temperature updated", slog.String("current_temperature", currentTemp), slog.String("name", c.Device.Name))

			mode := c.parseMode(v.Port1.Mode)
			if v.Port1.Power == 0 {
				mode = "off"
			}
			token = c.mqtt.Publish(c.ModeStateTopic, 0, false, mode)
			if token.Error() != nil {
				slog.ErrorContext(ctx, "failed to publish ac mode state", slog.Any("error", token.Error()))
			}
			slog.InfoContext(ctx, "mode updated", slog.String("mode", mode), slog.String("name", c.Device.Name))

			swingMode := c.parseSwing(v.Port1.VSwing)
			token = c.mqtt.Publish(c.SwingModeStateTopic, 0, false, swingMode)
			if token.Error() != nil {
				slog.ErrorContext(ctx, "failed to publish ac swing mode state", slog.Any("error", token.Error()))
			}
			slog.InfoContext(ctx, "swing mode updated", slog.String("swing_mode", swingMode), slog.String("name", c.Device.Name))

			targetTemp := strconv.FormatFloat(v.Port1.Temperature, 'f', -1, 64)
			token = c.mqtt.Publish(c.TemperatureStateTopic, 0, false, targetTemp)
			if token.Error() != nil {
				slog.ErrorContext(ctx, "failed to publish ac target temperature state", slog.Any("error", token.Error()))
			}
			slog.InfoContext(ctx, "target temperature updated", slog.String("target_temperature", targetTemp), slog.String("name", c.Device.Name))
		}
	}()
}

func (c *Climate) CommandSubscriptions() {
	fanMode := c.mqtt.Subscribe(c.FanModeCommandTopic, 0, c.handleFanMode)
	go func() {
		_ = fanMode.Wait()
		if fanMode.Error() != nil {
			slog.Error("error subscribing", slog.Any("error", fanMode.Error()))
		} else {
			slog.Info("subscribed to topic", slog.String("topic", c.FanModeCommandTopic))
		}
	}()

	operationMode := c.mqtt.Subscribe(c.ModeCommandTopic, 0, c.handleMode)
	go func() {
		_ = operationMode.Wait()
		if operationMode.Error() != nil {
			slog.Error("error subscribing", slog.Any("error", operationMode.Error()))
		} else {
			slog.Info("subscribed to topic", slog.String("topic", c.ModeCommandTopic))
		}
	}()

	targetTemp := c.mqtt.Subscribe(c.TemperatureCommandTopic, 0, c.handleTargetTemp)
	go func() {
		_ = targetTemp.Wait()
		if targetTemp.Error() != nil {
			slog.Error("error subscribing", slog.Any("error", targetTemp.Error()))
		} else {
			slog.Info("subscribed to topic", slog.String("topic", c.TemperatureCommandTopic))
		}
	}()

	swingMode := c.mqtt.Subscribe(c.SwingModeCommandTopic, 0, c.handleswingMode)
	go func() {
		_ = swingMode.Wait()
		if swingMode.Error() != nil {
			slog.Error("error subscribing", slog.Any("error", swingMode.Error()))
		} else {
			slog.Info("subscribed to topic", slog.String("topic", c.SwingModeCommandTopic))
		}
	}()
}

func (c *Climate) PublishDiscovery() {
	payload, err := json.Marshal(c)
	if err != nil {
		slog.Error("failed to marshal payload")
	}

	token := c.mqtt.Publish(c.DiscoveryTopic(), 0, true, payload)
	go func() {
		_ = token.Wait()
		if token.Error() != nil {
			slog.Error("failed to publish discovery", slog.String("device", c.Device.Name), slog.Any("error", token.Error()))
		}
	}()
}

func (c *Climate) parseSwing(i int) string {
	switch i {
	case 0:
		return "off"
	case 1:
		return "on"
	default:
		return ""
	}
}

func (c *Climate) parseFanMode(f daikin.Fan) string {
	switch f {
	case 3, 4:
		return "low"
	case 5, 6:
		return "medium"
	case 7:
		return "high"
	case 17:
		return "auto"
	default:
		return ""
	}
}

func (c *Climate) parseMode(m daikin.Mode) string {
	switch m {
	case 0:
		return "auto"
	case 2:
		return "dry"
	case 3:
		return "cool"
	case 4:
		return "heat"
	case 6:
		return "fan_only"
	default:
		return m.String()
	}
}

func (c *Climate) DiscoveryPayload() []byte {
	payload, err := json.Marshal(c)
	if err != nil {
		slog.Error("failed to marshal payload")
		return []byte{}
	}

	return payload
}

func (c *Climate) DiscoveryTopic() string {
	return fmt.Sprintf("homeassistant/climate/%s/config", c.UniqueId)
}

func (c *Climate) handleFanMode(_ pahomqtt.Client, msg pahomqtt.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	translate := map[string]daikin.Fan{
		"auto":   17,
		"low":    3,
		"medium": 5,
		"high":   7,
	}
	value, ok := translate[string(msg.Payload())]
	if !ok {
		slog.Error("unknown value for fan mode", slog.Int("value", int(value)))
		return
	}
	slog.Debug("set fan mode received", slog.String("value", value.String()))
	_, err := c.daikinClient.SetState(ctx, daikin.DesiredState{
		Port1: daikin.PortState{
			Fan: &value,
		},
	})
	if err != nil {
		slog.Error("failed to send fan mode to ac", slog.String("name", c.Device.Name))
	}
}

func (c *Climate) handleMode(_ pahomqtt.Client, msg pahomqtt.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	payload := string(msg.Payload())
	var desiredState = daikin.DesiredState{}
	modesMap := map[string]daikin.Mode{
		"auto":     0,
		"dry":      2,
		"cool":     3,
		"heat":     4,
		"fan_only": 6,
	}

	mode, ok := modesMap[payload]
	if ok {
		desiredState.Port1.Mode = &mode
		desiredState.Port1.Power = intPtr(1)
	} else if payload == "off" {
		desiredState.Port1.Power = intPtr(0)
	} else {
		slog.Error("unknown mode", slog.String("value", payload))
		return
	}
	slog.Debug("set fan mode received", slog.String("value", mode.String()))
	_, err := c.daikinClient.SetState(ctx, desiredState)
	if err != nil {
		slog.Error("failed to send mode to ac", slog.String("name", c.Device.Name), slog.Any("error", err))
	}
}

func (c *Climate) handleTargetTemp(_ pahomqtt.Client, msg pahomqtt.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	payload := string(msg.Payload())
	slog.Debug("set temperature received", slog.String("value", payload))

	targetTemp, err := strconv.ParseFloat(payload, 64)
	if err != nil {
		slog.ErrorContext(ctx, "string to float conversion failed", slog.Any("error", err))
		return
	}

	_, err = c.daikinClient.SetState(ctx, daikin.DesiredState{
		Port1: daikin.PortState{
			Temperature: &targetTemp,
		},
	})
	if err != nil {
		slog.Error("failed to send temperature to ac", slog.String("name", c.Device.Name), slog.Any("error", err))
	}
}

func (c *Climate) handleswingMode(_ pahomqtt.Client, msg pahomqtt.Message) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	payload := string(msg.Payload())
	slog.Debug("set swing mode received", slog.String("value", payload))

	swingMap := map[string]int{
		"off": 0,
		"on":  1,
	}
	value, ok := swingMap[payload]
	if !ok {
		slog.Error("unknown swing mode value", slog.Int("value", value))
		return
	}
	_, err := c.daikinClient.SetState(ctx, daikin.DesiredState{
		Port1: daikin.PortState{
			VSwing: intPtr(value),
		},
	})
	if err != nil {
		slog.Error("failed to send temperature to ac", slog.String("name", c.Device.Name), slog.Any("error", err))
	}
}

func intPtr(i int) *int {
	return &i
}
