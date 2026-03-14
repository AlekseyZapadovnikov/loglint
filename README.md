# loglint

`loglint` проверяет лог-сообщения для `log/slog` и `go.uber.org/zap`.

Проверяемые правила:
- `lowercase` — сообщение должно начинаться со строчной буквы
- `english` — сообщение должно содержать только английский текст
- `symbols` — сообщение не должно содержать шумные символы и эмодзи
- `sensitive` — сообщение и structured fields не должны содержать потенциально чувствительные данные

По умолчанию включены все 4 правила.

## Содержание

- [Quick Start](#quick-start)
- [Конфигурация](#конфигурация)
- [Что проверяется](#что-проверяется)
- [Поддержка логгеров](#поддержка-логгеров)
- [Sensitive keywords](#sensitive-keywords)
- [Ограничения](#ограничения)
- [Пример диагностики](#пример-диагностики)
- [Тесты](#тесты)
- [License](#license)

## Quick Start

Основной способ использования — готовый `custom-gcl` из релизов.

1. Скачай `custom-gcl` из раздела Releases для своей платформы.

2. Сделай бинарник исполняемым:

```bash
chmod +x ./custom-gcl
```

3. Положи бинарник рядом с `.golangci.yml` и добавь настройку:

```yaml
version: "2"

linters:
  default: none
  enable:
    - loglint
  settings:
    custom:
      loglint:
        type: module
        description: Checks slog and zap log messages for style and possible sensitive data.
        settings:
          enabled_rules:
            - all
          disabled_rules:
            - symbols
```

4. Запусти скачанный бинарник:

```bash
./custom-gcl run ./...
```

### Standalone (dev/debug)

`loglint` публикуется как дополнительный standalone-бинарник для локальной отладки.

```bash
./loglint ./...
```

## Конфигурация

Для `golangci-lint` используется схема:

```yaml
enabled_rules:
  - all

disabled_rules: []

sensitive:
  extra_keywords:
    - client_secret
    - private_key
  replace_defaults: false
```

### Параметры

- `enabled_rules` — если не задан, включаются все правила; если задан, стартуем только с него
- `disabled_rules` — вычитает правила после `enabled_rules`
- `sensitive.extra_keywords` — дополнительные sensitive keywords
- `sensitive.replace_defaults` — если `true`, заменяет дефолтный sensitive-словарь

### Допустимые значения

- `all`
- `lowercase`
- `english`
- `symbols`
- `sensitive`

### Семантика

- `all` разрешён только в `enabled_rules`
- `all` в `disabled_rules` — ошибка конфигурации
- неизвестное правило — ошибка конфигурации
- текущая конфигурация `sensitive` backward compatible

### Примеры

Включить только часть правил:

```yaml
enabled_rules:
  - lowercase
  - sensitive
```

Включить всё, кроме `symbols`:

```yaml
enabled_rules:
  - all
disabled_rules:
  - symbols
```

Оставить только `sensitive` и добавить свои keywords:

```yaml
enabled_rules:
  - sensitive

sensitive:
  extra_keywords:
    - client_secret
    - private_key
```

Полностью заменить дефолтные sensitive keywords:

```yaml
sensitive:
  extra_keywords:
    - my_secret
    - api_token
  replace_defaults: true
```

## Что проверяется

### Lowercase

```go
// bad
slog.Info("Starting server")

// good
slog.Info("starting server")
```

### English

```go
// bad
slog.Info("запуск сервера")

// good
slog.Info("server started")
```

### Symbols

```go
// bad
slog.Info("request started!")
slog.Info("user logged in 🎉")

// good
slog.Info("request started")
```

### Sensitive

```go
// bad
slog.Info("password: " + password)
slog.Info("user logged in", "session_id", sessionID)

// good
slog.Info("password changed successfully")
slog.Info("user logged in", "user_id", userID)
```

## Поддержка логгеров

### `log/slog`

Поддерживаются:
- `slog.Debug`, `slog.Info`, `slog.Warn`, `slog.Error`
- `logger.Debug`, `logger.Info`, `logger.Warn`, `logger.Error`
- поля в форме `"key", value`

Пока не поддерживаются:
- `slog.Attr`
- `slog.LogValue`

### `go.uber.org/zap`

Поддерживаются:
- `(*zap.Logger).Debug`, `Info`, `Warn`, `Error`
- `(*zap.SugaredLogger).Debugw`, `Infow`, `Warnw`, `Errorw`
- `zap.String("key", value)` и другие field constructors
- для sugared logger: `"key", value`

Пока не поддерживаются:
- `DPanic`, `Panic`, `Fatal`
- sugared-методы без суффикса `w`

## Sensitive keywords

По умолчанию используются:

- `password`, `passwd`, `pwd`
- `secret`
- `token`, `api_key`, `apikey`, `access_token`, `refresh_token`
- `authorization`, `bearer`
- `cookie`, `session`, `session_id`
- `credential`, `credentials`

Keywords нормализуются:
- регистр игнорируется
- `client_secret`, `client-secret`, `client secret` считаются одинаковыми
- пустые строки и дубли отбрасываются

## Ограничения

- анализируются только Go-пакеты, переданные на вход
- для `sensitive` используется keyword matching, а не regex engine
- для `sensitive` статические сообщения проверяются по key-value паттернам (`:`/`=`),
  а для динамических сообщений дополнительно анализируется статический префикс

## Пример диагностики

```text
internal/handler/auth.go:42:9: log message must start with a lowercase letter
```

## Тесты

```bash
go test ./...
```

## License

MIT
