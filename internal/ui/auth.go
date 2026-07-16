package ui

import (
	"context"
	"html/template"
	"io"
	"strings"

	"github.com/a-h/templ"
)

type AuthPage struct {
	AppName  string
	AppURL   string
	Mode     string
	Title    string
	Subtitle string
	Error    string
	Notice   string
}

func AuthDocument(vm AuthPage) templ.Component {
	return templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, authHTML(vm))
		return err
	})
}

func authHTML(vm AuthPage) string {
	if vm.Title == "" {
		vm.Title = "Login"
	}
	var b strings.Builder
	b.WriteString(`<!doctype html><html lang="id"><head><meta charset="utf-8">`)
	b.WriteString(`<meta name="viewport" content="width=device-width, initial-scale=1">`)
	b.WriteString(`<title>` + template.HTMLEscapeString(vm.Title+" - "+vm.AppName) + `</title>`)
	b.WriteString(`<meta name="description" content="Portal operasional Taiwan Education Consulting.">`)
	b.WriteString(`<meta name="theme-color" content="#071225">`)
	b.WriteString(`<link rel="icon" href="/favicon.svg" type="image/svg+xml">`)
	b.WriteString(`<link rel="stylesheet" href="/assets/app.css?v=` + assetVersion + `">`)
	b.WriteString(`<link rel="stylesheet" href="/assets/icons.css?v=` + assetVersion + `">`)
	b.WriteString(`<script src="/assets/app.js?v=` + assetVersion + `" defer></script>`)
	b.WriteString(`</head><body><main class="auth-shell"><section class="auth-panel">`)
	b.WriteString(`<div class="auth-brand"><div class="brand-mark" aria-hidden="true"><span></span></div><div><p>TAIWAN</p><strong>Education Consulting</strong></div></div>`)
	b.WriteString(`<h1>` + template.HTMLEscapeString(vm.Title) + `</h1><p>` + template.HTMLEscapeString(vm.Subtitle) + `</p>`)
	if vm.Error != "" {
		b.WriteString(`<div class="auth-error">` + template.HTMLEscapeString(vm.Error) + `</div>`)
	}
	if vm.Notice != "" {
		b.WriteString(`<div class="auth-notice">` + template.HTMLEscapeString(vm.Notice) + `</div>`)
	}
	if vm.Mode == "register" {
		b.WriteString(registerForm())
		b.WriteString(`<p class="auth-switch">Sudah punya akun? <a href="/login">Login</a></p>`)
	} else {
		b.WriteString(loginForm())
		b.WriteString(demoAccounts())
	}
	b.WriteString(`</section><aside class="auth-side"><div><span>Portal operasional</span><h2>Owner, staff, dan client masuk dari satu pintu.</h2><p>Akses akun dikelola oleh owner dan staff.</p></div></aside></main></body></html>`)
	return b.String()
}

func loginForm() string {
	return `<form method="post" action="/login" class="auth-form"><label>Username<input name="username" autocomplete="username" required autofocus></label><label>Password<input name="password" type="password" autocomplete="current-password" required></label><button class="primary-button" type="submit">Login</button></form>`
}

func registerForm() string {
	return `<form method="post" action="/register" class="auth-form"><label>Nama lengkap<input name="name" autocomplete="name" required autofocus></label><label>Username<input name="username" autocomplete="username" required></label><label>Email<input name="email" type="email" autocomplete="email"></label><label>No. WhatsApp<input name="phone" autocomplete="tel"></label><label>Password<input name="password" type="password" autocomplete="new-password" minlength="8" required></label><button class="primary-button" type="submit">Register Client</button></form>`
}

func demoAccounts() string {
	var b strings.Builder
	b.WriteString(`<div class="demo-accounts"><p class="demo-accounts-title">Akun Demo</p>`)
	for _, acc := range []struct{ role, label, username, password string }{
		{"owner", "Owner", "owner", "owner12345"},
		{"staff", "Staff", "staff", "staff12345"},
		{"student", "Client", "student", "student12345"},
	} {
		b.WriteString(`<button type="button" class="demo-account" data-demo-username="` + template.HTMLEscapeString(acc.username) + `" data-demo-password="` + template.HTMLEscapeString(acc.password) + `">`)
		b.WriteString(`<span class="demo-account-role">` + template.HTMLEscapeString(acc.label) + `</span>`)
		b.WriteString(`<span class="demo-account-cred">` + template.HTMLEscapeString(acc.username) + ` / ` + template.HTMLEscapeString(acc.password) + `</span>`)
		b.WriteString(`</button>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}
