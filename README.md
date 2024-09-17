# wb-tech-level-0
Для корректной работы приложения (а также тестов), необходим .env файл с нужными переменными.

Пример .env содержимого файла есть в .env-example

С корректным .env файлом запуск производится через docker compose:

```docker compose up```

После этого, интерфейс доступен по адресу localhost:8080

# Testing 
При тестировании используется [testcontainers](https://github.com/testcontainers/testcontainers-go), поэтому нет необходимости отдельно запускать kafka и postgres, достаточно лишь запустить тестирование, например:


```go test -v```
