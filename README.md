# i18n

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/flamego/i18n/Go?logo=github&style=for-the-badge)](https://github.com/flamego/i18n/actions?query=workflow%3AGo)
[![GoDoc](https://img.shields.io/badge/GoDoc-Reference-blue?style=for-the-badge&logo=go)](https://pkg.go.dev/github.com/flamego/i18n?tab=doc)

Package i18n is a middleware that provides internationalization and localization for [Flamego](https://github.com/flamego/flamego).

## Installation

```zsh
go get github.com/flamego/i18n
```

## Getting started

```ini
# locales/locale_en-US.ini
greeting = How are you?
```

```ini
# locales/locale_zh-CN.ini
greeting = 你好吗？
```

```go
package main

import (
	"github.com/flamego/flamego"
	"github.com/flamego/i18n"
)

func main() {
	f := flamego.Classic()
	f.Use(i18n.I18n(
		i18n.Options{
			Languages: []i18n.Language{
				{Name: "en-US", Description: "English"},
				{Name: "zh-CN", Description: "简体中文"},
			},
		},
	))
	f.Get("/", func(l i18n.Locale) {
		message := l.Translate("greeting")
		// ...
	})
	f.Run()
}
```

## Getting help

- Read [documentation and examples](https://flamego.dev/middleware/i18n.html).
- Please [file an issue](https://github.com/flamego/flamego/issues) or [start a discussion](https://github.com/flamego/flamego/discussions) on the [flamego/flamego](https://github.com/flamego/flamego) repository.

## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.
