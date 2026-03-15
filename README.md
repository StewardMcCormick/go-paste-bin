# 🎓 Учебный проект: API для Paste-Bin-like сервиса
- Сервис для временного хранения и распространения сниппетов текста
- Учебный проект для изучения и практики Clean Architecture и System Design
- Основная функциональность:
  - Создание сниппета с указанным временем жизни, политикой доступа(public, protected, private) и паролем, если сниппет 
    указан как protected
  - Получение сниппета по сгенерированному на стороне сервера хэшу
  - Обновление сниппета

- Технические реализации:
  - Clean Architecture с разделением на слои
  - RESTful-API
  - Кэш с использованием Redis
  - Rate limiting с использованием Redis и алгоритма Token Bucket 
  - Многоуровневая система доступа (public/private/protected) с API Key аутентификацией и bcrypt хешированием
  - Оптимизация работы с PostgreSQL: batch-update и удаление истекших записей с помощью фоновых worker-ов
  - Инфраструктура: Docker, docker-compose, миграции (golang-migrate), graceful shutdown, structured logging (zap), 
    Makefile
  - Unit + интеграционные тесты (testcontainers)
  
## 👨‍💻 Технологические стэк
- Язык: Go 1.25+
- HTTP-роутер: Chi
- Базы данных: PostgreSQL (+ миграции с использованием golang-migrate)
- Кэширование и Rate Limiting: Redis (+ Lua для атомарной записи в хранилище Rate Limiter`а)
- Конфигурация: godotenv
- Логирование: uber-go/zap
- Тестирование: testify, mockery, testcontainers
- Контейнеризация: Docker + Docker-Compose

## 🏛️ Архитектура проекта
Сервис спроектирован с использованием принципов **Clean Architecture**:

```
go-paste-bin/
├── 📂cmd/
│    └── 📂app/                        # точка входа main
├── 📂config/                          # конфигурации компонентов
├── 📂internal/                        
│    ├── 📂adapter/                    # компоненты взаимодействия с внешними сервисами
│    │    ├── 📂postgres/              
│    │    └── 📂redis/                 
│    ├── 📂app/                        # код инициализации компонентов, запуска и остановки приложения
│    ├── 📂domain/                     # доменные модели
│    ├── 📂dto/                        # транспортные модели
│    ├── 📂error/                      # компоненты для управления ошибками
│    ├── 📂handler/                    # http-handlers
│    │    ├── 📂middleware/             
│    ├── 📂migrations/                 # файлы мигарций .sql
│    ├── 📂repository/                 # репозиторный слой
│    ├── 📂usecase/                    # серивисный слой
│    └── 📂util/                       # дополнительные утилиты
│        ├── 📂app_context/            # взаимодействие с контекстом приложения
│        ├── 📂db_clean_up_worker/     # утилита фоновой очистки устаревших сниппетов и API-ключей 
│        ├── 📂http_util/              # утилита-обортка http.ResponseWriter для логирования 
│        ├── 📂rate/                   # Rate Limitting-утилита
│        ├── 📂security/               # утилита безопасности для хэширования и проверки хэшей
│        ├── 📂validation/             # утилита валидации
│        └── 📂views_worker/           # утилита фонового batch-update статистики просмотров сниппетов
├── 📂pkg/                             
│   ├── 📂httpserver/                  # компонент конфигурации и запуска http-сервера
│   ├── 📂logging/                     # компонент конфигурации и запуска логера
│   └── 📂render/                      # компонент сериализации объектов в JSON и записи в http.ResponseWritter
├── ⚙️config.yaml                      # конфигурация приложения
├── ⚙️.env                             # secrets для подключения к внешним сервисам в local-окружении 
├── ⚙️.env.docker-single               # secrets для подключения к внешним сервисам в docker-контейнере
├── ⚙️.env.docker-compose              # secrets для подключения к внешним сервисам в docker-compose
├── ⚙️docker-compose.yaml              # конфигурация docker-compose
├── ⚙️Dockerfile                       # конфигурация docker image
├── ⚙️Makefile                         # автоматизация сборки и запуска
```

## 🚀 Инструкция по запуску
#### Примечание: 
Подробную информацию по доступным командам `make` можно получить через `make help`

### Требования:
- Docker и Docker Compose - рекомендуется<br>
ИЛИ
- Go 1.25+, PostgreSQL, Redis (два инстанса для Cache и Rate Limiting) - для локального запуска<br>
ИЛИ
- Docker, PostgreSQL, Redis (два инстанса для Cache и Rate Limiting) - для запуска сервиса внутри docker-контейнера

### Запуск через Docker-Compose
1. Клонировать репозиторий: `git clone https://github.com/StewardMcCormick/go-paste-bin.git`
2. Перейти в папку с проектом: `cd ./go-paste-bin`
3. Скопировать `.env.docker-compose.example` в `.env.docker-compose` (можно оставить дефолтные значение)
4. Выполнить `make run-docker-compose`
5. Сервис будет доступен по адресу `http://localhost:8080`

### Запуск локально
1. Клонировать репозиторий: `git clone https://github.com/StewardMcCormick/go-paste-bin.git`
2. Перейти в папку с проектом: `cd ./go-paste-bin`
3. Скопировать `.env.example` в `.env` и заполнить значениями для свои PostgreSQL и Redis (запустить их, если это не сделано)
4. Выполнить `make run-local`
5. Сервис будет доступен по адресу `http://localhost:8080`

### Запуск в Docker-контейнере
1. Клонировать репозиторий: `git clone https://github.com/StewardMcCormick/go-paste-bin.git`
2. Перейти в папку с проектом: `cd ./go-paste-bin`
3. Скопировать `.env.docker-single.example` в `.env.docker-single` и заполнить значениями для свои PostgreSQL и Redis(запустить их, если это не сделано)
4. Выполнить `make run-docker-single`

## 📞 API endpoints

### ✅ Регистрация и login

#### Регистрация

```bash
curl -X POST `http://localhost:8080/registration` \
  -H "Content-Type: application/json" \
  -d `{ "username": "name",
        "password": "password"
      }`
```

**Ответы**:
- `201 Created`: с созданным пользователем, его id и API-Ключом
- `400 BadRequest`: при невалидном теле запроса (неправильный JSON или ошибка валидации - в этом случае содержит информацию о неправильно заполненных полях)
- `409 Conflict`: если пользователь с таким username уже существует

#### Login - обновления API-ключа (с удалением старого)

```bash
curl -X POST `http://localhost:8080/login` \
  -H "Content-Type: application/json" \
  -H "X-API-Key": "{user-api-key}"
  -d '{ "username": "name",
        "password": "password"
      }'
```

**Ответы**:
- `201 Created`: с новым API-ключом
- `400 BadRequest`: при невалидном теле запроса(неправильный JSON или ошибка валидации - в этом случае содержит информацию о неправильно заполненных полях)
- `404 Not Found`: если пользователь с таким username не найден
- `401 Unauthorized`: если пользователь прислал неверный пароль/API-ключ

### 💬 Работа со сниппетами:
#### Примечание: во всех запросах ниже сервер может вернуть `401 Unauthorized` при невалидном API-ключе

#### Создание сниппета:

```bash
curl -X POST `http://localhost:8080/api/v1/paste` \
  -H "Content-Type: application/json" \
  -H "X-API-Key": "{user-api-key}"
  -d '{ "content": "content",
        "privacy": "public"                 # одно из "public", "protected", "private",
        "Password": "password",             # если "privacy" установленно в "protected"
        "expire_at": "2023-10-21T15:04:05Z" #  RFC 3339 / ISO 8601 строка, если не передается - заполняется дефолтным значением (см. `config.yaml`)
      }'
```

**Ответ**:
- `201 created`: с созданным сниппетом, его id и заголовком Location
- `400 BadRequest`: при невалидном теле запроса (неправильный JSON или ошибка валидации - в этом случае содержит информацию о неправильно заполненных полях)

#### Получение сниппета(public или private):
```bash
curl -X GET `http://localhost:8080/api/v1/paste/{paste_hash}` \
  -H "Content-Type: application/json" \
  -H "X-API-Key": "{user-api-key}"
```

**Ответы**:
- `200 OK`: с информацией о сниппете
- `403 Forbidden`: если пользователь запрашивает `private` сниппет, который ему не принадлежит
- `404 Not Found`: если сниппет не был найден

#### Получение сниппета(protected):
```bash
curl -X GET `http://localhost:8080/api/v1/paste/{paste_hash}` \
  -H "Content-Type: application/json" \
  -H "X-API-Key": "{user-api-key}" \
  -d '{
        "password": "password"
      }'
```

**Ответ**:
- `200 OK`: с информацией о сниппете
- `401 Unauthorized`: если пароль неверный
- `404 Not Found`: если сниппет не был найден
- `400 BadRequest`: при невалидном теле запроса (неправильный JSON или ошибка валидации - в этом случае содержит информацию о неправильно заполненных полях)

#### Обновление сниппета
```bash
curl -X PATCH `http://localhost:8080/api/v1/paste/{paste_hash}` \
  -H "Content-Type: application/json" \
  -H "X-API-Key": "{user-api-key}"
  -d '{ "content": "content",
        "privacy": "public"                 # одно из "public", "protected", "private",
        "Password": "password",             # если "privacy" установленно в "protected"
        "expire_at": "2023-10-21T15:04:05Z" #  RFC 3339 / ISO 8601 строка, если не передается - заполняется дефолтным значением (см. `config.yaml`)
      }'
```
#### Примечание: все поля в этом запросе опциональны, кроме случая, когда `privacy` устанавливается в `protected` - тогда поле `password` обязательно

**Ответы**:
- `200 OK`: с обновленной информацией и сниппете
- `400 BadRequest`: при невалидном теле запроса (неправильный JSON или ошибка валидации - в этом случае содержит информацию о неправильно заполненных полях)
- `404 Not Found`: если сниппет не был найден


## 📑 Тестирование:
- **Unit-тест** покрывают всю бизнес логику(coverage > 80%) с использованием сгенерированных с помощью mockery моков
- **Интеграционные тесты** покрывают взаимодействие репозиторного слоя с внещними БД (поднимаются тестовые инстансы с помощью testcontainers)


### Запуск тестов:
- `make test`: запуск всех тестов
- `make test-cover`: запуск всех тестов с отчетом по покрытию

## ☎️ Контакты:
- Email: bessonoven@mail.ru
- Telegram: @bessonov_en
