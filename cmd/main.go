package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/alan-b-lima/almodon/internal/api/v1"
	"github.com/alan-b-lima/almodon/internal/middleware"

	"github.com/alan-b-lima/ansi-escape-sequences"
)

var StdOut = os.Stdout

func main() {
	log := middleware.NewLogger(StdOut, "")
	style := Styles()

	ln, err := net.Listen("tcp", ":4545")
	if err != nil {
		log.Println(err)
		return
	}

	url := "http://" + strings.Replace(ln.Addr().String(), "[::]", "localhost", 1)
	log.Printf("Server listening at %s\n", style.HyperLink(url))

	var mux http.ServeMux

	api, err := api.New()
	if err != nil {
		log.Println(err)
		return
	}
	defer api.Close()

	mux.Handle("/", http.FileServer(http.Dir("../ui/web/")))
	mux.Handle("/api/", api)
	mux.HandleFunc("/terminate/{timeout}", Terminate)

	srv := http.Server{Handler: TrafficLogMiddleware(log, style, &mux)}

	done := EnableGracefulShutdown(func() {
		log.Println("Shutting server down...")
		srv.Shutdown(context.Background())
	})

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		log.Println(err)
	}

	<-done
}

func TrafficLogMiddleware(log *middleware.Logger, s Style, handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := middleware.NewW(w)

		handler.ServeHTTP(rw, r)

		pen := s.StatusCodePen(rw.StatusCode())

		var barr [32]byte
		b := bytes.NewBuffer(barr[:0])
		pen.Writer = b

		io.WriteString(&pen, fmt.Sprintf(" %03d ", rw.StatusCode()))
		log.Printf("%s %s %s %s\n", b.String(), r.RemoteAddr, r.Method, r.URL)
	}
}

var Signals chan<- os.Signal

func EnableGracefulShutdown(fn func()) <-chan struct{} {
	signals := make(chan os.Signal, 1)
	Signals = signals

	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	done := make(chan struct{}, 1)

	go func() {
		<-signals
		fn()
		done <- struct{}{}
	}()

	return done
}

type Style struct {
	HyperLink func(string) string

	Pens    map[int]ansi.Pen
	NoStyle ansi.Pen

	Enabled bool
}

func Styles() (s Style) {
	s.Enabled = ansi.EnableVirtualTerminal(os.Stdout.Fd()) == nil
	s.NoStyle.SetStyle(false)

	if s.Enabled {
		var Success ansi.Pen
		var Redirect ansi.Pen
		var ClientError ansi.Pen
		var ServerError ansi.Pen

		Success.BGColor(ansi.RGBFromHex(0x0ed145))
		Success.FGColor(ansi.RGBFromHex(0xffffff))

		Redirect.BGColor(ansi.RGBFromHex(0x4b53cc))
		Redirect.FGColor(ansi.RGBFromHex(0xffffff))

		ClientError.BGColor(ansi.RGBFromHex(0xea1d1d))
		ClientError.FGColor(ansi.RGBFromHex(0xffffff))

		ServerError.BGColor(ansi.RGBFromHex(0x88001b))
		ServerError.FGColor(ansi.RGBFromHex(0xffffff))

		s.Pens = map[int]ansi.Pen{
			2: Success,
			3: Redirect,
			4: ClientError,
			5: ServerError,
		}
		s.HyperLink = hyperlink

		return s
	}

	s.HyperLink = func(s string) string { return s }
	s.Pens = map[int]ansi.Pen{}

	return s
}

func (s *Style) StatusCodePen(status int) ansi.Pen {
	pen, in := s.Pens[status/100]
	if !in {
		return s.NoStyle
	}

	return pen
}

func hyperlink(link string) string {
	var pen ansi.Pen
	pen.FGColor(ansi.RGBFromHex(0x4e8597))

	return pen.Sprint(ansi.HyperLinkP(link))
}

func Terminate(w http.ResponseWriter, r *http.Request) {
	ms, err := strconv.Atoi(r.PathValue("timeout"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	go func() {
		time.Sleep(time.Duration(ms) * time.Millisecond)
		Signals <- syscall.SIGTERM
	}()

	w.WriteHeader(http.StatusNoContent)
}
