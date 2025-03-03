# Setup

- make State() request to check if ok and grab info.
- publish mqtt discovery messages for each sensor and entity
  - climate
  - external temp sensor (with availability topic for when the AC is off)
  - econo switch
  - confort switch
  - powerchill switch
  - good_sleep switch
- grab and update ac current state async
- subscribe to command topics
  - handle commands
    - call SetState()
    - publish result for each sensor/entity

## Discovery climate

```json
{
  "name": "Ar condicionado",
  "unique_id": "DAIKIN46ACB4",
  "current_temperature_topic": "daikin/DAIKIN46ACB4/temperature",
  "fan_modes": ["auto", "low", "high", "medium"],
  "fan_mode_command_topic": "daikin/DAIKIN46ACB4/fan_mode/set",
  "fan_mode_state_topic": "daikin/DAIKIN46ACB4/fan_mode/state",
  "device": {
    "manufacturer": "Daikin Brazil",
    "identifiers": "DAIKIN46ACB4",
    "name": "Daikin Ecoswing Smart R-32 Suíte"
  }
}
```

# Switch economy

```json
{
  "name": "Econômico",
  "unique_id": "DAIKIN000001_Economy",
  "command_topic": "daikinsmartac/DAIKIN000001/economy/set",
  "state_topic": "daikinsmartac/DAIKIN000001/economy/state",
  "state_off ": "OFF",
  "state_on ": "ON",
  "payload_off ": "OFF",
  "payload_on ": "ON",
  "device": {
    "manufacturer": "Daikin Brasil",
    "via_device": "OneControl Gateway",
    "identifiers": "hvac_0000000DA7271001",
    "name": "Daikin Smart AC Brasil Bridge",
    "sw_version": "1.0"
  }
}
```

# state update

- fan
- operation
- temperature
- switches
