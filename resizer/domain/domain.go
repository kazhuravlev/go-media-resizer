package domain

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

var (
	BadResponse           = errors.New("bad response code")
	CannotFetch           = errors.New("cannot fetch source image")
	ErrInvalidSourceImage = errors.New("invalid source image")
)

type Domain struct {
	httpClient *http.Client
	pool       sync.Pool
	resizer    Resizer
}

func New(httpClient *http.Client, resizer Resizer) (*Domain, error) {
	bufferPool := sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	return &Domain{
		httpClient: httpClient,
		pool:       bufferPool,
		resizer:    resizer,
	}, nil
}

func (d *Domain) FetchResize(ctx context.Context, buf *bytes.Buffer, url string, w, h int) (string, error) {
	bufSrc := d.pool.Get().(*bytes.Buffer)
	bufSrc.Reset()
	defer d.pool.Put(bufSrc)

	ct, err := d.fetch(ctx, url, bufSrc)
	if err != nil {
		return "", CannotFetch
	}

	if err := d.resizer.Resize(ct, bufSrc, buf, w, h); err != nil {
		return "", err
	}

	return ct, nil
}

func (d *Domain) fetch(ctx context.Context, url string, buf *bytes.Buffer) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	resp, err := d.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	defer func() {
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", err
	}

	// небезопасно доверять это заголовкам, оставлено для простоты имплементации
	contentType := strings.ToLower(resp.Header.Get("content-type"))

	if _, err := io.Copy(buf, resp.Body); err != nil {
		return "", err
	}

	return contentType, nil
}
