# Стандарт разработки на Go 1.27

Используйте Go 1.27.3 из development image с digest pin и `go 1.27.3` в
`go.mod`. `GOTOOLCHAIN=local` запрещает незаметную загрузку другого compiler
в контейнере. CI читает `go.mod` либо использует тот же точный base image.

Сохраните действующие соглашения provider: context первым аргументом,
framework diagnostics, `tflog`, проверку TLS, идентификаторы crypto/rand,
конкретные типы и race checks. [Руководство Go 1.26](../standards/go-1.26-style.md)
фиксирует прежние соглашения; указанные там версии инструментов исторические.

Миграция SDK не требует новых возможностей языка. В Go 1.27 `go mod tidy`
объединяет повторяющиеся require blocks; фиксируйте фактический результат tidy.
Схема, state, import IDs и HCL остаются контрактом совместимости.
Версия framework во время этой миграции остаётся прежней, с точным pin.

Используйте golangci-lint v2.15.0 для анализа Go 1.27. Выполняйте `task all`
в development container. Live acceptance и публикация release остаются
отдельными проверками.

Источники: [Go 1.27 release notes](https://go.dev/doc/go1.27),
[Go toolchains](https://go.dev/doc/toolchain),
[golangci-lint v2.15.0](https://github.com/golangci/golangci-lint/releases/tag/v2.15.0).
