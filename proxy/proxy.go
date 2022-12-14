package proxy

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	"github.com/andrebq/auth/client"
)

var (
	loginTmpl = template.Must(template.New("__root__").Parse(`
{{ define "unavailable" }}
<!DOCTYPE html>
<html>
	<head>
		<title>Unavailable</title>
	</head>
	<body>
		{{ if .Error }}
		<p class="error">{{.Error}}</p>
		{{ else }}
		<p class="error">Site is not available at the moment</p>
		{{ end}}
	</body>
</html>
{{ end }}
{{ define "login" }}
<!DOCTYPE html>
<html>
	<head>
		<title>Please login</title>
	</head>
	<body>
		<form method="POST" action="./login">
			<fieldset>
				<caption>Login credentials</caption>
				<section>
					<label for="username">Username</label>
					<input id="username" name="username" placeholder="username"/>
				</section>
				<section>
					<label for="username">Password</label>
					<input id="password" name="password" placeholder="password" type="password"/>
				</section>
				<section>
					<button>Login</button>
				</section>
				{{ if .Error }}
				<p class="error">{{.Error}}</p>
				{{ end}}
			</fieldset>
		</form>
	</body>
</html>
{{ end }}
`))
)

func Handler(upstreamBase string, apiBase string) (http.Handler, error) {
	upstreamURL, err := url.Parse(upstreamBase)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	cli := client.New(apiBase)
	mux.Handle("/.auth/login", handleLoginUI(cli))
	mux.Handle("/", handleProxy(upstreamURL, cli))
	return mux, nil
}

func handleLoginUI(cli *client.C) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			renderLoginUI(w, r)
		case "POST":
			loginAndRedirect(w, r, cli)
		}
	})
}

func renderLoginUI(w http.ResponseWriter, req *http.Request) {
	renderTemplate(w, loginTmpl, http.StatusOK, "login", struct{ Error string }{})
}

func renderTemplate(w http.ResponseWriter, tmpl *template.Template, status int, name string, data interface{}) {
	buf := bytes.Buffer{}
	err := tmpl.ExecuteTemplate(&buf, name, data)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.Header().Add("Content-Length", strconv.Itoa(buf.Len()))
	// TODO: add proper csrf
	w.WriteHeader(status)
	w.Write(buf.Bytes())
}

func loginAndRedirect(w http.ResponseWriter, req *http.Request, cli *client.C) {
	err := req.ParseForm()
	if err != nil {
		renderTemplate(w, loginTmpl, http.StatusOK, "login", struct{ Error string }{Error: err.Error()})
		return
	}
	username := req.FormValue("username")
	password := req.FormValue("password")
	if len(username) == 0 || len(password) == 0 {
		renderTemplate(w, loginTmpl, http.StatusBadRequest, "login", struct{ Error string }{Error: "Please inform your username and password"})
		return
	}
	session, err := cli.StartSession(req.Context(), username, password, time.Hour*24)
	if err != nil {
		renderTemplate(w, loginTmpl, http.StatusInternalServerError, "unavailable", struct{ Error string }{Error: "Cannot perform authentication at the moment, please try again later"})
		return
	}
	cookie := http.Cookie{
		Name: "auth.session",
		Path: "/",
		// TODO: encrypt the cookie content
		Value:    session,
		Expires:  time.Now().Add(time.Hour * 24),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Domain:   req.URL.Hostname(),
		Secure:   true,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, req, "/", http.StatusSeeOther)
}

func handleProxy(upstreamBase *url.URL, cli *client.C) http.Handler {
	rev := httputil.NewSingleHostReverseProxy(upstreamBase)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth.session")
		if err != nil {
			redirectOrFail(w, r)
			return
		}
		if !validCookie(r.Context(), cookie, cli) {
			redirectOrFail(w, r)
			return
		}
		var upstreamCookies []*http.Cookie
		for _, v := range r.Cookies() {
			if v.Name == "auth.session" {
				continue
			}
			upstreamCookies = append(upstreamCookies, v)
		}
		r.Header.Del("Cookie")
		for _, c := range upstreamCookies {
			r.Header.Add("Cookie", c.String())
		}
		rev.ServeHTTP(w, r)
	})
}

func validCookie(ctx context.Context, c *http.Cookie, cli *client.C) bool {
	// TODO: encrypt this cookie
	_, _, err := cli.ValidateToken(ctx, c.Value)
	if err != nil {
		return false
	}
	return true
}

func redirectOrFail(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Please login and try again", http.StatusUnauthorized)
		return
	}
	loginURL := *req.URL
	loginURL.RawQuery = ""
	loginURL.Path = "/.auth/login"

	// Force a GET request to upstream
	http.Redirect(w, req, loginURL.String(), http.StatusSeeOther)
}
