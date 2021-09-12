// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package i18n

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flamego/flamego"

	"github.com/flamego/i18n/testdata/primary"
)

func TestI18n(t *testing.T) {
	t.Run("no language is specified", func(t *testing.T) {
		require.PanicsWithValue(t,
			"i18n: no language is specified",
			func() {
				f := flamego.NewWithLogger(&bytes.Buffer{})
				f.Use(I18n(Options{}))
			},
		)
	})

	t.Run("bad local directory", func(t *testing.T) {
		require.PanicsWithValue(t,
			"i18n: init locales: open from local: open 404/locale_en-US.ini: no such file or directory",
			func() {
				f := flamego.NewWithLogger(&bytes.Buffer{})
				f.Use(I18n(
					Options{
						Directory: "404",
						Languages: []Language{
							{Name: "en-US", Description: "English"},
						},
					},
				))
			},
		)
	})

	t.Run("bad FileSystem", func(t *testing.T) {
		require.PanicsWithValue(t,
			"i18n: init locales: open from FileSystem: open locale_it-IT.ini: file does not exist",
			func() {
				f := flamego.NewWithLogger(&bytes.Buffer{})
				f.Use(I18n(
					Options{
						FileSystem: http.FS(primary.Locales),
						Languages: []Language{
							{Name: "it-IT", Description: "Italiano"},
						},
					},
				))
			},
		)
	})

	t.Run("good", func(t *testing.T) {
		f := flamego.NewWithLogger(&bytes.Buffer{})
		f.Use(I18n(
			Options{
				Directory:         "testdata/primary",
				AppendDirectories: []string{"testdata/secondary"},
				Languages: []Language{
					{Name: "en-US", Description: "English"},
					{Name: "zh-CN", Description: "简体中文"},
				},
			},
		))
	})

	t.Run("pick by URL query parameter", func(t *testing.T) {
		f := flamego.NewWithLogger(&bytes.Buffer{})
		f.Use(I18n(
			Options{
				Directory:         "testdata/primary",
				AppendDirectories: []string{"testdata/secondary"},
				Languages: []Language{
					{Name: "en-US", Description: "English"},
					{Name: "zh-CN", Description: "简体中文"},
				},
			},
		))
		f.Get("/", func(l Locale) string {
			return l.Lang() + " " + l.Description()
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/?lang=zh-CN", nil)
		require.Nil(t, err)

		f.ServeHTTP(resp, req)

		require.Equal(t, "zh-CN 简体中文", resp.Body.String())
		require.Equal(t, "lang=zh-CN; Path=/; Max-Age=2147483647; HttpOnly; SameSite=Lax", resp.Header().Get("Set-Cookie"))
	})

	t.Run("pick by cookie", func(t *testing.T) {
		f := flamego.NewWithLogger(&bytes.Buffer{})
		f.Use(I18n(
			Options{
				Directory:         "testdata/primary",
				AppendDirectories: []string{"testdata/secondary"},
				Languages: []Language{
					{Name: "en-US", Description: "English"},
					{Name: "zh-CN", Description: "简体中文"},
				},
			},
		))
		f.Get("/", func(l Locale) string {
			return l.Lang() + " " + l.Description()
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.Nil(t, err)

		req.AddCookie(
			&http.Cookie{
				Name:  "lang",
				Value: "zh-CN",
			},
		)
		f.ServeHTTP(resp, req)

		require.Equal(t, "zh-CN 简体中文", resp.Body.String())
		require.Empty(t, resp.Header().Get("Set-Cookie"))
	})

	t.Run("pick by Accept-Language", func(t *testing.T) {
		f := flamego.NewWithLogger(&bytes.Buffer{})
		f.Use(I18n(
			Options{
				Directory:         "testdata/primary",
				AppendDirectories: []string{"testdata/secondary"},
				Languages: []Language{
					{Name: "en-US", Description: "English"},
					{Name: "zh-CN", Description: "简体中文"},
				},
			},
		))
		f.Get("/", func(l Locale) string {
			return l.Lang() + " " + l.Description()
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.Nil(t, err)

		req.Header.Add("Accept-Language", "zh")
		f.ServeHTTP(resp, req)

		require.Equal(t, "zh-CN 简体中文", resp.Body.String())
		require.Equal(t, "lang=zh-CN; Path=/; Max-Age=2147483647; HttpOnly; SameSite=Lax", resp.Header().Get("Set-Cookie"))
	})
}

func TestI18n_Translate(t *testing.T) {
	t.Run("load from local", func(t *testing.T) {
		f := flamego.NewWithLogger(&bytes.Buffer{})
		f.Use(I18n(
			Options{
				Directory:         "testdata/primary",
				AppendDirectories: []string{"testdata/secondary"},
				Languages: []Language{
					{Name: "en-US", Description: "English"},
					{Name: "zh-CN", Description: "简体中文"},
				},
			},
		))
		f.Get("/", func(l Locale) string {
			return l.Translate("greeting")
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.Nil(t, err)

		f.ServeHTTP(resp, req)

		require.Equal(t, "What's up?", resp.Body.String())
	})

	t.Run("load from FileSystem", func(t *testing.T) {
		f := flamego.NewWithLogger(&bytes.Buffer{})
		f.Use(I18n(
			Options{
				FileSystem:        http.FS(primary.Locales),
				AppendDirectories: []string{"testdata/secondary"},
				Languages: []Language{
					{Name: "en-US", Description: "English"},
					{Name: "zh-CN", Description: "简体中文"},
				},
			},
		))
		f.Get("/", func(l Locale) string {
			return l.Translate("greeting")
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		require.Nil(t, err)

		f.ServeHTTP(resp, req)

		require.Equal(t, "What's up?", resp.Body.String())
	})

	t.Run("fallback", func(t *testing.T) {
		f := flamego.NewWithLogger(&bytes.Buffer{})
		f.Use(I18n(
			Options{
				FileSystem:        http.FS(primary.Locales),
				AppendDirectories: []string{"testdata/secondary"},
				Languages: []Language{
					{Name: "en-US", Description: "English"},
					{Name: "zh-CN", Description: "简体中文"},
				},
			},
		))
		f.Get("/", func(l Locale) string {
			return l.Translate("greeting")
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/?lang=zh-CN", nil)
		require.Nil(t, err)

		f.ServeHTTP(resp, req)

		require.Equal(t, "What's up?", resp.Body.String())
		require.Equal(t, "lang=zh-CN; Path=/; Max-Age=2147483647; HttpOnly; SameSite=Lax", resp.Header().Get("Set-Cookie"))
	})

	t.Run("language not found", func(t *testing.T) {
		f := flamego.NewWithLogger(&bytes.Buffer{})
		f.Use(I18n(
			Options{
				Directory: "testdata/primary",
				Languages: []Language{
					{Name: "en-US", Description: "English"},
				},
			},
		))
		f.Get("/", func(l Locale) string {
			return l.Translate("greeting")
		})

		resp := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, "/?lang=zh-CN", nil)
		require.Nil(t, err)

		f.ServeHTTP(resp, req)

		require.Equal(t, "How are you?", resp.Body.String())
	})
}
