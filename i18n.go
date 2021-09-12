// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package i18n

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"reflect"

	"github.com/go-i18n/i18n"
	"github.com/pkg/errors"
	"golang.org/x/text/language"

	"github.com/flamego/flamego"
)

// CookieOptions contains options for setting HTTP cookies.
type CookieOptions struct {
	// Name is the name of the cookie. Default is "lang".
	Name string
	// Path is the Path attribute of the cookie. Default is "/".
	Path string
	// Domain is the Domain attribute of the cookie. Default is not set.
	Domain string
	// MaxAge is the MaxAge attribute of the cookie. Default is math.MaxInt.
	MaxAge int
	// Secure specifies whether to set Secure for the cookie.
	Secure bool
	// HTTPOnly specifies whether to set HTTPOnly for the cookie.
	HTTPOnly bool
	// SameSite is the SameSite attribute of the cookie. Default is
	// http.SameSiteLaxMode.
	SameSite http.SameSite
}

// Language contains the name and description of a language.
type Language struct {
	// Name is the BCP 47 language name, e.g. "en-US".
	Name string
	// Description is the descriptive name of the language, e.g. "English".
	Description string
}

// Options contains options for the i18n.I18n middleware.
type Options struct {
	// FileSystem is the interface for supporting any implementation of the
	// http.FileSystem.
	FileSystem http.FileSystem
	// Directory is the primary directory to load locale files. This value is
	// ignored when FileSystem is set. Default is "locales".
	Directory string
	// AppendDirectories is a list of additional directories to load locale files
	// for overwriting locale files that are loaded from FileSystem or Directory.
	AppendDirectories []string
	// Languages is the list of languages to load locale files for.
	Languages []Language
	// Default is the name of the default language to fall back for missing
	// translations. Default is the first element of Languages.
	Default string
	// NameFormat is the name format of locale files. Default is "locale_%s.ini".
	NameFormat string
	// QueryParameter is the name of the URL query parameter to accept language
	// override. Default is "lang".
	QueryParameter string
	// Cookie is a set of options for setting HTTP cookies.
	Cookie CookieOptions
}

// Locale is the message translator of a language.
type Locale interface {
	// Lang returns the BCP 47 language name of the locale.
	Lang() string
	// Description returns the descriptive name of the locale.
	Description() string
	// Translate translates the message of the given key. It falls back to use the
	// "Default" to translate if the given key does not exist in the current locale.
	Translate(key string, args ...interface{}) string
}

type locale struct {
	fallback *i18n.Locale
	current  *i18n.Locale
}

func (l *locale) Lang() string {
	return l.current.Lang()
}

func (l *locale) Description() string {
	return l.current.Description()
}

func (l *locale) Translate(key string, args ...interface{}) string {
	return l.current.TranslateWithFallback(l.fallback, key, args...)
}

// isFile returns true if given path exists as a file (i.e. not a directory).
func isFile(path string) bool {
	f, e := os.Stat(path)
	if e != nil {
		return false
	}
	return !f.IsDir()
}

// initLocales initializes a locale store with list of provided languages
// loading from http.FileSystem and/or local files. If both `fs` and `dir` are
// provided, only `fs` is considered.
func initLocales(langs []Language, nameFormat string, fs http.FileSystem, dir string, others ...string) (*i18n.Store, language.Matcher, error) {
	s := i18n.NewStore()

	tags := make([]language.Tag, 0, len(langs))
	for _, lang := range langs {
		tag, err := language.Parse(lang.Name)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "parse %q", lang.Name)
		}
		tags = append(tags, tag)

		filename := fmt.Sprintf(nameFormat, lang.Name)
		var source io.ReadCloser
		if fs != nil {
			source, err = fs.Open(filename)
			if err != nil {
				return nil, nil, errors.Wrap(err, "open from FileSystem")
			}
		} else {
			source, err = os.Open(path.Join(dir, filename))
			if err != nil {
				return nil, nil, errors.Wrap(err, "open from local")
			}
		}

		otherSources := make([]interface{}, 0, len(others))
		for _, other := range others {
			otherpath := path.Join(other, filename)
			if !isFile(otherpath) {
				continue
			}
			otherSources = append(otherSources, otherpath)
		}

		_, err = s.AddLocale(lang.Name, lang.Description, source, otherSources...)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "add locale for %q", lang.Name)
		}
	}

	return s, language.NewMatcher(tags), nil
}

// I18n returns a middleware handler that injects i18n.Locale into the request
// context, which is used for localization.
func I18n(opt Options) flamego.Handler {
	parseOptions := func(opts Options) Options {
		if opts.Directory == "" {
			opts.Directory = "locales"
		}

		if len(opts.Languages) == 0 {
			panic("i18n: no language is specified")
		}

		if opts.Default == "" {
			opts.Default = opts.Languages[0].Name
		}

		if opts.NameFormat == "" {
			opts.NameFormat = "locale_%s.ini"
		}

		if opts.QueryParameter == "" {
			opts.QueryParameter = "lang"
		}

		if reflect.DeepEqual(opts.Cookie, CookieOptions{}) {
			opts.Cookie = CookieOptions{
				HTTPOnly: true,
			}
		}
		if opts.Cookie.Name == "" {
			opts.Cookie.Name = "lang"
		}
		if opts.Cookie.SameSite < http.SameSiteDefaultMode || opts.Cookie.SameSite > http.SameSiteNoneMode {
			opts.Cookie.SameSite = http.SameSiteLaxMode
		}
		if opts.Cookie.Path == "" {
			opts.Cookie.Path = "/"
		}
		if opts.Cookie.MaxAge <= 0 {
			// TODO: math.MaxInt is only available since Go 1.17, should start using it once
			//  Go 1.17 becomes the minimum required version.
			opts.Cookie.MaxAge = 1<<31 - 1 // = 2147483647
		}

		return opts
	}

	opt = parseOptions(opt)

	store, matcher, err := initLocales(opt.Languages, opt.NameFormat, opt.FileSystem, opt.Directory, opt.AppendDirectories...)
	if err != nil {
		panic("i18n: init locales: " + err.Error())
	}

	fallback, err := store.Locale(opt.Default)
	if err != nil {
		panic("i18n: get fallback: " + err.Error())
	}

	return flamego.ContextInvoker(func(c flamego.Context) {
		setCookie := func(lang string) {
			c.SetCookie(
				http.Cookie{
					Name:     opt.Cookie.Name,
					Value:    lang,
					Path:     opt.Cookie.Path,
					Domain:   opt.Cookie.Domain,
					MaxAge:   opt.Cookie.MaxAge,
					Secure:   opt.Cookie.Secure,
					HttpOnly: opt.Cookie.HTTPOnly,
					SameSite: opt.Cookie.SameSite,
				},
			)
		}

		// 1. Check URL query parameter
		lang := c.Query(opt.QueryParameter)
		if lang != "" {
			setCookie(lang)
		}

		// 2. Check cookie
		if lang == "" {
			lang = c.Cookie(opt.Cookie.Name)
		}

		// 3. Check the first element in the "Accept-Language" header
		if lang == "" {
			tags, _, _ := language.ParseAcceptLanguage(c.Request().Header.Get("Accept-Language"))
			tag, _, confidence := matcher.Match(tags...)
			if confidence != language.No {
				lang = tag.String()
				setCookie(lang)
			}
		}

		// 4. Fall back to default
		if lang == "" {
			lang = opt.Default
			setCookie(lang)
		}

		var l Locale
		current, err := store.Locale(lang)
		if err != nil {
			if err != i18n.ErrLocalNotFound {
				panic("i18n: get locale: " + err.Error())
			}

			l = fallback
		} else {
			l = &locale{
				fallback: fallback,
				current:  current,
			}
		}
		c.MapTo(l, (*Locale)(nil))
	})
}
