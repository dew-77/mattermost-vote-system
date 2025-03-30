# mattermost-vote-system
Функционал системы голосования для мессенджера Mattermost в формате бота


## Описание

Этот проект реализует бота для системы голосования внутри чатов мессенджера Mattermost. Бот позволяет пользователям создавать голосования, голосовать за предложенные варианты, просматривать результаты, завершать голосования досрочно и удалять их.

### Основные возможности:
- **Создание голосования**: Бот регистрирует голосование и возвращает сообщение с ID голосования и вариантами ответов.
- **Голосование**: Пользователи могут отправить команду, указывая ID голосования и вариант ответа.
- **Просмотр результатов**: Любой пользователь может запросить текущие результаты голосования.
- **Завершение голосования**: Создатель голосования может завершить его досрочно.
- **Удаление голосования**: Возможность удаления голосования.

## Структура проекта

Проект состоит из следующих ключевых частей:


```bash
mattermost-vote-system/
├── cmd/ # Основной запуск бота
│   └── bot/
│       └── main.go # Точка входа для бота
├── internal/ # Логика приложения
│   ├── app/
│   │   ├── app.go # Главная логика приложения
│   │   └── handlers.go # Обработчики команд
│   ├── config/ # Конфигурационные файлы
│   │   └── config.go # Чтение и обработка конфигураций
│   ├── models/ # Модели данных
│   │   ├── poll.go # Модель голосования
│   │   └── vote.go # Модель для голосов
│   ├── repository/ # Работа с данными
│   │   ├── tarantool.go # Репозиторий для работы с Tarantool
│   │   └── repository.go # # Абстракция репозитория
│   └── mattermost/ # Взаимодействие с Mattermost API
│       └── client.go # Клиент для общения с Mattermost
├── pkg/ # Вспомогательные пакеты
│   └── logger/ # Логирование
│       └── logger.go # Реализация логирования
├── docker/ # Docker конфигурации
│   ├── bot/
│   │   └── Dockerfile # Dockerfile для бота
│   └── tarantool/ 
│       ├── Dockerfile # Dockerfile для Tarantool
│       └── init.lua # Инициализация базы данных Tarantool
├── docker-compose.yml # Docker Compose конфигурация
├── go.mod # Модульные зависимости Go
├── go.sum # Контроль версий зависимостей
├── .env.example # Файл переменных среды
├── config.yaml # Конфигурационный файл
└── README.md 
```

## Установка


0. Разверните Mattermost локально - [подробнее тут](https://docs.mattermost.com/install/install-docker.html#)

1. В настройках Mattermost включите поддержку Webhooks:

```plain
System Console → Developer → Enable WebSocket connections → ✅ Enabled
```

2. В разделе интеграций создайте нового бота:
```plain
System Console → Integrations → Bots → Create new bot
```

3. Клонируйте репозиторий:

    ```bash
    git clone https://github.com/yourusername/mattermost-voting-bot.git
    cd mattermost-voting-bot
    ```


4. В корне репозитория создайте новый файл .env и заполните его по примеру .env.example:
```env
MATTERMOST_TOKEN=<токен бота>
MATTERMOST_TEAMNAME=<название команды>
MATTERMOST_BOTUSERID=<id бота>
```

5. Заполните `config.yaml` в корне репозитория:
```yaml
mattermost:
  serverURL: "http://host.docker.internal:8065"
  token: "<токен бота>"
  teamName: "<название команды>"
  botUserID: "<id бота>"

tarantool:
  host: "tarantool"
  port: 3301
  user: "admin"
  password: "<пароль БД>"
  space: "polls"

bot:
  logLevel: "info"
```

6. Соберите контейнеры Docker:

    ```bash
    docker-compose up --build
    ```

7. После сборки и запуска контейнеров, бот будет доступен и подключен к системе Mattermost.

8. Добавьте бота в нужный чат


## Использование

Список доступных команд можно получить, если использовать команду:
```bash
@имя_бота help
```
Или написать в личные сообщения бота.

Примеры команд:
- **create "Заголовок" "Вариант 1" "Вариант 2" ...** - Создать новое голосование
- **vote [ID голосования] [номер варианта]** - Проголосовать за вариант
- **results [ID голосования]** - Показать результаты голосования
- **finish [ID голосования]** - Завершить голосование (только для создателя)
- **delete [ID голосования]** - Удалить голосование (только для создателя)
- **help** - Показать справку`