package domain

import (
	"bytes"
	"context"
	"crypto/sha1"
	"errors"
	bolt "go.etcd.io/bbolt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	ErrFetchImage = errors.New("cannot fetch source image")
	ErrHasNoCache = errors.New("cannot fetch image in cache")
)

var (
	bucketContentType = []byte("ct")
	bucketFile        = []byte("files")
)

type Domain struct {
	httpClient *http.Client
	pool       sync.Pool
	resizer    Resizer
	db         *bolt.DB
}

func New(httpClient *http.Client, resizer Resizer, dbFilepath string) (*Domain, error) {
	bufferPool := sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}

	db, err := bolt.Open(dbFilepath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	errCreate := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(bucketFile); err != nil {
			return err
		}

		if _, err := tx.CreateBucketIfNotExists(bucketContentType); err != nil {
			return err
		}

		return nil
	})
	if errCreate != nil {
		return nil, errCreate
	}

	return &Domain{
		httpClient: httpClient,
		pool:       bufferPool,
		resizer:    resizer,
		db:         db,
	}, nil
}

func (d *Domain) FetchResize(ctx context.Context, buf *bytes.Buffer, url string, w, h int) (string, error) {
	bufSrc := d.pool.Get().(*bytes.Buffer)
	bufSrc.Reset()
	defer d.pool.Put(bufSrc)

	ct, err := d.fetch(ctx, url, bufSrc)
	if err != nil {
		log.Println(err)
		return "", ErrFetchImage
	}

	if err := d.resizer.Resize(ct, bufSrc, buf, w, h); err != nil {
		return "", err
	}

	return ct, nil
}

func (d *Domain) fetch(ctx context.Context, url string, buf *bytes.Buffer) (string, error) {
	var contentType string

	h := sha1.New()
	h.Write([]byte(url))
	key := h.Sum(nil)

	err := d.db.View(func(tx *bolt.Tx) error {
		ct := tx.Bucket(bucketContentType).Get(key)
		if ct == nil {
			return ErrHasNoCache
		}

		file := tx.Bucket(bucketFile).Get(key)
		if file == nil {
			return ErrHasNoCache
		}

		contentType = string(ct)
		buf.Write(file)

		return nil
	})

	if err == nil {
		return contentType, nil
	}

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
	contentType = strings.ToLower(resp.Header.Get("content-type"))

	if _, err := io.Copy(buf, resp.Body); err != nil {
		return "", err
	}

	errUpd := d.db.Update(func(tx *bolt.Tx) error {
		if err := tx.Bucket(bucketContentType).Put(key, []byte(contentType)); err != nil {
			return err
		}

		if err := tx.Bucket(bucketFile).Put(key, buf.Bytes()); err != nil {
			return err
		}

		return nil
	})
	if errUpd != nil {
		log.Printf("Cannot insert: %s", errUpd)
	}

	return contentType, nil
}
