## Вопросы для разогрева

---

1. Опишите самую интересную задачу в программировании, которую вам приходилось решать?

Можно сказать, что этот проект и был самой интересной задачей которую мне приходилось решить в программировании. Требовалось достаточно большое кол-во функционала в довольно сжатые сроки, а также задача такого рода, была для меня новой.

2. Расскажите о своем самом большом факапе? Что вы предприняли для решения проблемы?

Самый большой факап был на групповом проекте в рамках предмета в университете. Мы неправильно оценили объем работы и наши возможности, из-за чего сильно не успевали к сроку. Пришлось много работать в последние 2 ночи перед сдачей. После этого я понял, что нужно лучше продумывать свою работу, трезво оценивать ее объем и свои силы, а также брать время с запасом.

3. Каковы ваши ожидания от участия в буткемпе?

Меня очень привлекает возможность участия в буткемпе, потому что cloud.ru большая компания и занимается реализацией интересных проектов. Я уверен, что у меня получится хорошо вырасти как по hard, так и по soft скилам, а также мне было бы очень интересно прийти в офис, пообщаться с коллегами и приобщиться к культуре компании.

## Сборка и запуск проекта

---

#### Перед сборкой установи зависимости

- Go 1.21+
- Docker
- PostgreSQL

### Сборка
#### 1. Настрой файл конфигурации

Создать в корневом каталоге файл config.yaml и заполнить все поля, как описано в примере example_config.yaml.

#### 2. Создай Docker контейнер c БД

- выполни команду docker run -d \
  --name <container_name> \
  -e POSTGRES_USER=<db_user> \
  -e POSTGRES_PASSWORD=<db_password> \
  -e POSTGRES_DB=<db_name> \
  -p 5432:5432 \
  postgres

  Переменные db_user, db_password и db_name взять из конфигурации, вместо container_name написать свое имя контейнера. После запуска БД будет доступна по адресу localhost:5432.

#### 3. Примени миграции к базе данных

Выполни команду migrate -source file://./internal/migrations -database "postgres://<db_user>:<db_pqssword>@localhost:5432/<db_name>?sslmode=disable" up
Не забудь заменить все переменные в <> на собственные.

### Запуск
#### Выполни следующие команды

- docker start <container_name>
- go run cmd/main.go