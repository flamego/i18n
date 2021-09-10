// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package i18n

import (
	"net/http"
	"reflect"

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
	Name        string
	Description string
}

// Options contains options for the i18n.I18n middleware.
type Options struct {
	// Directory is the primary directory to load locale files. This value is
	// ignored when FileSystem is set. Default is "locales".
	Directory string
	// AppendDirectories is a list of additional directories to load locale files
	// for overwriting locale files that are loaded from Directory. This value is
	// ignored when FileSystem is set.
	AppendDirectories []string
	// FileSystem is the interface for supporting any implementation of the
	// http.FileSystem.
	FileSystem http.FileSystem
	// Languages is the list of languages to load locale files for.
	Languages []Language
	// Default is the name of the default language to fall back for missing
	// translations. Default is none.
	Default string
	// NameFormat is the name format of locale files. Default is "locale_%s.ini".
	NameFormat string
	// URLParameter is the name of the URL parameter to accept language override.
	// Default is "lang".
	URLParameter string
	// Cookie is a set of options for setting HTTP cookies.
	Cookie CookieOptions
}

// todo
type Locale struct {
}

// I18n returns a middleware handler that injects i18n.Locale into the request
// context, which is used for localization.
func I18n(opts ...Options) flamego.Handler {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	parseOptions := func(opts Options) Options {
		if opts.Directory == "" {
			opts.Directory = "locales"
		}

		if opts.FileSystem == nil {
			// todo: init new file system from opts.Directory and opts.AppendDirectories
		}

		if opts.NameFormat == "" {
			opts.NameFormat = "locale_%s.ini"
		}

		if opts.URLParameter == "" {
			opts.URLParameter = "lang"
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
			opts.Cookie.MaxAge = 2 ^ 31 - 1 // = 2147483647 = 2038-01-19 04:14:07
		}

		return opts
	}

	opt = parseOptions(opt)

	return flamego.ContextInvoker(func(c flamego.Context) {

	})
}
