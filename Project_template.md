## Изучите [README.md](.\README.md) файл и структуру проекта.

# Задание 1

1. Спроектировал to be архитектуру КиноБездны, разделив всю систему на отдельные домены и организовав интеграционное взаимодействие и единую точку вызова сервисов.
Результат предоставил в виде контейнерной диаграммы в нотации С4.
[C4 диаграмма](/src/docs/c4Conteiner.puml)

# Задание 2

### 1. Proxy
Реализовал бесшовный переход с применением паттерна Strangler Fig в части реализации прокси-сервиса (API Gateway), с помощью которого можно будет постепенно переключать траффик, используя фиче-флаг.

Реализовал сервис в ./src/microservices/proxy.
Конфигурация для запуска сервиса через docker-compose уже добавлена

- После реализации запустил postman тесты - они все должны были зеленые (кроме events).
- Отправил запросы к API Gateway:
   ```bash
   curl http://localhost:8000/api/movies
   ```
- Протестировал постепенный переход, изменив переменную окружения MOVIES_MIGRATION_PERCENT в файле docker-compose.yml.

### 2. Kafka
Для этого нужно сделал MVP сервис events, который будет при вызове API создавать и сам же читать сообщения в топике Kafka.

    - Разработал сервис с consumer'ами и producer'ами.
    - Разработал простой API, при вызове которого будут создаваться события User/Payment/Movie и обрабатываться внутри сервиса с записью в лог
    - Добавил в docker-compose новый сервис.

Необходимые тесты для проверки этого API вызываются при запуске npm run test:local из папки tests/postman 
Cкриншот тестов
[Скриншот тестов консоль](/src/docs/2TestResult.PNG)
[Скриншот тестов Postman](/src/docs/2TestResultPostman.PNG)

Cкриншот состояния топиков Kafka из UI http://localhost:8090 
[Cкриншот топиков Kafka](/src/docs/2KafkaTopicsState.PNG)

# Задание 3

Команда начала переезд в Kubernetes для лучшего масштабирования и повышения надежности. 
Необходимо было:
 - реализовать CI/CD для сборки прокси сервиса
 - реализовать необходимые конфигурационные файлы для переключения трафика.

### CI/CD

 В папке .github/worflows доработал деплой новых сервисов proxy и events в docker-build-push.yml, чтобы api-tests при сборке отрабатывали корректно при отправке коммита в мой репозиторий.

 После многочисленных попыток публикации образы пявились в репозитории github.

### Proxy в Kubernetes

#### Шаг 1
Для деплоя в kubernetes настроил параметры безопасности

#### Шаг 2

  Доработал src/kubernetes/event-service.yaml и src/kubernetes/proxy-service.yaml

  - Создал Deployment и Service 
  - Доработал ingress.yaml, чтобы можно было с помощью тестов проверить создание событий
  - Выполнил дальшейшие шаги для поднятия кластера:

  1. Создал namespace:
  ```bash
  kubectl apply -f src/kubernetes/namespace.yaml
  ```
  2. Создал секреты и переменные
  ```bash
  kubectl apply -f src/kubernetes/configmap.yaml
  kubectl apply -f src/kubernetes/secret.yaml
  kubectl apply -f src/kubernetes/dockerconfigsecret.yaml
  kubectl apply -f src/kubernetes/postgres-init-configmap.yaml
  ```

  3. Развернул базу данных:
  ```bash
  kubectl apply -f src/kubernetes/postgres.yaml
  ```
  4. Развернул Kafka:
  ```bash
  kubectl apply -f src/kubernetes/kafka/kafka.yaml
  ```

  5. Развернул монолит:
  ```bash
  kubectl apply -f src/kubernetes/monolith.yaml
  ```
  6. Развернул микросервисы:
  ```bash
  kubectl apply -f src/kubernetes/movies-service.yaml
  kubectl apply -f src/kubernetes/events-service.yaml
  ```
  7. Развернул прокси-сервис:
  ```bash
  kubectl apply -f src/kubernetes/proxy-service.yaml
  ```

  8. Добавил ingress

  - Добавил аддон
  ```bash
  minikube addons enable ingress
  ```
  ```bash
  kubectl apply -f src/kubernetes/ingress.yaml
  ```
  9. Добавил в /etc/hosts
  127.0.0.1 cinemaabyss.example.com

  10. Вызвал
  ```bash
  minikube tunnel
  ```
  11. Вызвал https://cinemaabyss.example.com/api/movies
  Увидел вывод списка фильмов

  12. Запустил тесты из папки tests/postman
  ```bash
   npm run test:kubernetes
  ```
#### Шаг 3  

Cкриншот тестов для cinemaabyss.example.com после  npm run test:kubernetes
[Скриншот тестов консоль](/src/docs/2TestResult.PNG)

 и  скриншот вывода event-service после вызова тестов.
[Скриншот логов](/src/docs/3TestResult_event-service.PNG)

# Задание 4
Для простоты дальнейшего обновления и развертывания реализовал helm-чарты для прокси-сервиса и проверил работу 

Для этого:
1. Директорию helm отредактировал файл values.yaml

- Вместо ghcr.io/db-exp/cinemaabysstest/proxy-service записал свой путь до образа для всех сервисов
- для imagePullSecret проставил свое значение imagePullSecrets:dockerconfigjson 
  

2. В папке ./templates/services заполнил шаблоны для proxy-service.yaml и events-service.yaml (сделал шаблоны для быстрого обновления и установки)


3. Проверка установки
Сначала удалим установку руками

```bash
kubectl delete all --all -n cinemaabyss
kubectl delete  namespace cinemaabyss
```
Запуск
```bash
helm install cinemaabyss .\src\kubernetes\helm --namespace cinemaabyss --create-namespace
```

Проверка развертывания:
```bash
kubectl get pods -n cinemaabyss
minikube tunnel
```
Результат развертывания helm
[Скриншот helm](/src/docs/4HelmResult.PNG)

Результат вызова https://cinemaabyss.example.com/api/movies
[Скриншот helm](/src/docs/4cinemaabyssResult.PNG)


