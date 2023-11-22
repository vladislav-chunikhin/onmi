## Technical assignment

- [Task](docs/golang_test.md)

### Content

- [Makefile to run unit-tests and linter](Makefile)
    - ``make test-unit``
    - ``make lint-go``
- [GitHub Workflow](.github/workflows/test-onmi.yaml)

### Improvements

- Получение конфигураций из yaml файла
- Реализация самого внешнего сервиса
- Использование Circuit Breaker паттерна для минимизации каскадного сбоя, если говорим про микросервисы
- Контейнеризация приложения