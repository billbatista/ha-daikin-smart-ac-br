# Agradecimentos

Este projeto só foi possível graças ao trabalho feito pelo [crossworth](https://github.com/crossworth) no repositório [daikin](https://github.com/crossworth/daikin/).

# Para quem é

Pessoas que possuem o ar condicionado da Daikin modelo Split Ecoswing R-32 e possivelmente a variante Gold.

# Compatibilidade

A lista de aparelhos compatíveis deve ser a mesma presente [aqui](https://github.com/crossworth/daikin?tab=readme-ov-file#compatibilidade), mas eu não tenho como validar diretamente. Se você conseguiu utilizar, abra uma issue informando o modelo da unidade interna do seu aparelho para que possamos adicionar a lista.

# Como utilizar

## Secret key

Antes de começar, é necessário obter a `secret key` de cada ar condicionado que você possui. Se você ainda não configurou seus aparelhos no aplicativo `Daikin Smart AC Brasil`, faça agora seguindo as [intruções do fabricante](https://www.daikin.com.br/static/website/pdf/20210629%20-%20Manual%20do%20usu%C3%A1rio%20Daikin%20Smart%20AC.PDF).

Com o aplicativo configurado e funcionando, para obter a secret key é necessário acessar este [site](https://daikin-extract-secret-key.fly.dev/) (mantido pelo crossworth), e inserir seu usuário e senha do **aplicativo**. Ele irá retornar com uma lista de aparelhos e a secret key de cada um.

## Configuração

Faça uma cópia do arquivo `copy_example.yaml` e renomeie para `config.yaml`. Substitua as informações de acordo com a sua infraestrutura (usuário e senha do mqtt, IP do ar condicionado etc).

- unique_id (**obrigatório**): precisa ser um id único, que não se repita na sua instalação do Home Assistant. Ex.: `daikinsuite0001`
- address (**obrigatório**): o ip mais porta do ar condicionado. Ex.: `http://192.168.0.15:15914`
- secret_key (**obrigatório**): a chave obtida pelo site no passo acima
- operation_modes (**opcional**): estes são os modos suportados pelo seu aparelho, como `automático`, `desumidificador`, `aquecer` etc. Se não informado, a lista padrão será utilizada: `auto`, `off`, `cool`, `heat`, `dry`, `fan_only`. Se o seu modelo é apenas frio, passe a lista apenas com os demais modos:

```yaml
operation_modes:
  - auto
  - off
  - cool
  - dry
  - fan_only
```

- fan_modes (**opcional**): modos de ventilação do aparelho. Se não informado, a lista padrão é utilizada: `auto`, `low`, `medium`, `high`

Detalhes sobre os modos de ventilação: a lista é baseada nos modos suportados pelo Home Assistant. O aparelho de ar condicionado em si suporta mais modos. Alguns foram agrupados (baixo e média-baixa: low) e outros ainda precisam ser implementados, como o modo silencioso.

# Como executar

Este serviço pode ser executado de qualquer lugar da sua rede interna, desde que tenha acesso ao seu servidor MQTT e aos aparelhos de ar condicionado.

## Docker

Exemplo de como executar com o docker, passando o arquivo de configuração:
`docker run -v ./config.yaml:/app/config.yaml ghcr.io/billbatista/ha-daikin-smart-ac-br:latest`

## Executável (em breve)

Você pode baixar o executável de acordo com o seu sistema na página de [releases](). Com ele em mãos, no mesmo diretório crie o arquivo `config.yaml` conforme acima, e execute o programa.

# To do

- validação de configuração
- modo turbo
- modo economia
- modo conforto
- fan mode silencioso
- sensor de temperatura externa
- possibilitar uso de ssl e certificados na configuração do MQTT
- onboard mais fácil, fazendo a busca da secret key informando apenas o usuário e senha, como é feito no [site](https://daikin-extract-secret-key.fly.dev/)
- desabilitar discovery por uma interface web
