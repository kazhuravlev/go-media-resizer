package facade

import (
	"bytes"
	"context"
	"github.com/kazhuravlev/go-media-resizer/resizer/domain"
	"github.com/valyala/fasthttp"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
)

type Facade struct {
	log      *log.Logger
	hostPort string
	domain   *domain.Domain
	pool     sync.Pool
}

func New(logger *log.Logger, host, port string, d *domain.Domain) (*Facade, error) {
	bufferPool := sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	return &Facade{
		log:      logger,
		hostPort: net.JoinHostPort(host, port),
		domain:   d,
		pool:     bufferPool,
	}, nil
}

func (f *Facade) Run(ctx context.Context) {
	go f.run(ctx)
}

func (f *Facade) handleResize(ctx *fasthttp.RequestCtx) {
	url := string(ctx.QueryArgs().Peek("url"))
	w := string(ctx.QueryArgs().Peek("w"))
	h := string(ctx.QueryArgs().Peek("h"))

	wInt, err := strconv.Atoi(w)
	if err != nil {
		f.log.Printf("Cannot parse request: %s", err)
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}

	hInt, err := strconv.Atoi(h)
	if err != nil {
		f.log.Printf("Cannot parse request: %s", err)
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}

	buf := f.pool.Get().(*bytes.Buffer)
	buf.Reset()
	defer f.pool.Put(buf)

	ct, err := f.domain.FetchResize(ctx, buf, url, wInt, hInt)
	if err != nil {
		f.log.Printf("Cannot resize image: %s", err)
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}

	if _, err := buf.WriteTo(ctx); err != nil {
		f.log.Printf("Cannot write response: %s", err)
		ctx.SetStatusCode(http.StatusInternalServerError)
		return
	}

	ctx.SetContentType(ct)
	ctx.Response.Header.Set("Cache-Control", "max-age=3600")
}

func (f *Facade) run(ctx context.Context) {
	if err := fasthttp.ListenAndServe(f.hostPort, f.handleResize); err != nil {
		f.log.Fatalf("Error in ListenAndServe: %s", err)
	}
}
