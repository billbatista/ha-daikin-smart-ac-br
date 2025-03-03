# Agradecimentos

Este projeto só foi possível graças ao trabalho feito pelo [crossworth](https://github.com/crossworth) no repositório [daikin](https://github.com/crossworth/daikin/).

# Para quem é

Pessoas que possuem o ar condicionado da Daikin modelo Split Ecoswing R-32 e possivelmente a variante Gold.

# Como utilizar

## Secret key

Antes de começar, é necessário obter a `secret key` de cada ar condicionado que você possui. Se você ainda não configurou seus aparelhos no aplicativo `Daikin Smart AC Brasil`, faça agora seguindo as [intruções do fabricante](https://www.daikin.com.br/static/website/pdf/20210629%20-%20Manual%20do%20usu%C3%A1rio%20Daikin%20Smart%20AC.PDF).

Com o aplicativo configurado e funcionando, para obter a secret key é necessário acessar este [site](https://daikin-extract-secret-key.fly.dev/) (mantido pelo crossworth), e inserir seu usuário e senha do **aplicativo**. Ele irá retornar com uma lista de aparelhos e a secret key de cada um.

## Configuração

Faça uma cópia do arquivo `copy_example.yaml` e renomeie para `config.yaml`. Substitua as informações de acordo com a sua infraestrutura (usuário e senha do mqtt, IP do ar condicionado etc).

-

## Docker

## Executável

Você pode baixar o executável de acordo com o seu sistema na página de [releases]().

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
