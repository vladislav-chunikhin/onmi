## Technical assignment

- [Task](docs/golang_test.md)

### Content

- [Makefile to run unit-tests and linter](Makefile)
    - ``make test-unit``
    - ``make lint-go``
- [GitHub Workflow](.github/workflows/test-linter-onmi.yaml)

### Improvements

- Использование Circuit Breaker паттерна для минимизации каскадного сбоя, если говорим про микросервисы
- Контейнеризация приложения