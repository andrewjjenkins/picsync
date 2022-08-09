package googlephotos

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/google/uuid"
)

type CodeCatcher struct {
	Server     http.Server
	CatcherURL string
	State      string
	Codes      chan string
	Errors     chan error
}

const (
	port             = 8081
	CodeAcceptedPage = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Picsync: Google Photos Authorization Complete</title>
    <link href="http://fonts.googleapis.com/css?family=Open+Sans" rel="stylesheet" type="text/css">
  </head>
  <body>
    <h1>Login Complete</h1>
	<p>You have successfully authorized Picsync to read Google Photos</p>
	<p>You can return to the console now to begin syncing</p>
  </body>
</html>
<html>

`
)

func newCatcherHttpHandler(cc *CodeCatcher) http.HandlerFunc {
	catcherRequestError := func(w http.ResponseWriter, err error) {
		cc.Errors <- err

		// This is shared only over localhost, so no analysis of sensitivity of
		// error strings has been done.
		http.Error(w, err.Error(), 400)
	}

	catcherHttpHandler := func(w http.ResponseWriter, r *http.Request) {
		u := r.URL
		if u.Scheme != "http" && u.Scheme != "https" && u.Scheme != "" {
			catcherRequestError(w, fmt.Errorf("unknown scheme %s", u.Scheme))
			return
		}
		if r.Host != cc.Server.Addr {
			catcherRequestError(w, fmt.Errorf("unexpected host %s", r.Host))
			return
		}
		if u.Path == "/favicon.ico" {
			http.NotFound(w, r)
			return
		}
		if u.Path != "/picsyncCatchToken" {
			catcherRequestError(w, fmt.Errorf("unexpected path %s", u.Path))
			return
		}
		if r.Method != "GET" {
			catcherRequestError(w, fmt.Errorf("unexpected method %s", r.Method))
			return
		}
		if r.ContentLength != 0 {
			catcherRequestError(w, fmt.Errorf("unexpected non-zero content-length %d", r.ContentLength))
			return
		}
		params, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			catcherRequestError(w, fmt.Errorf("parsing query: %s", err.Error()))
			return
		}
		if len(params["state"]) != 1 {
			catcherRequestError(w, fmt.Errorf("redirect did not contain 1 state entry (replay prevention)"))
			return
		}
		state := params["state"][0]
		if state != cc.State {
			catcherRequestError(
				w,
				fmt.Errorf(
					"unexpected state mismatch (replay prevention), got \"%s\", expected \"%s\"",
					params["state"],
					cc.State,
				),
			)
			return
		}
		if len(params["code"]) != 1 {
			catcherRequestError(w, fmt.Errorf("redirect did not contain 1 code"))
		}
		code := params["code"][0]

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(200)
		fmt.Fprintf(w, "%s", CodeAcceptedPage)

		cc.Codes <- code
	}
	return catcherHttpHandler
}

func newCodeCatcher() (*CodeCatcher, error) {
	serverAddr := fmt.Sprintf("127.0.0.1:%d", port)
	catcherUrl := fmt.Sprintf("http://%s/picsyncCatchToken", serverAddr)

	catcher := &CodeCatcher{
		Codes:  make(chan string, 100),
		Errors: make(chan error, 100),
		Server: http.Server{
			Addr: serverAddr,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			}),
		},
		CatcherURL: catcherUrl,
		State:      uuid.NewString(),
	}

	catcherHandler := newCatcherHttpHandler(catcher)
	catcher.Server.Handler = catcherHandler

	listener, err := net.Listen("tcp", catcher.Server.Addr)
	if err != nil {
		return nil, err
	}
	go catcher.Server.Serve(listener)

	return catcher, nil
}
