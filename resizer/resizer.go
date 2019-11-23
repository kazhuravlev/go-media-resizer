package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kazhuravlev/sample-media-resizer/resizer/domain"
	"github.com/kazhuravlev/sample-media-resizer/resizer/facade"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	host, port string
	dbFilename string
)

func init() {
	flag.StringVar(&host, "host", "127.0.0.1", "")
	flag.StringVar(&port, "port", "9000", "")
	flag.StringVar(&dbFilename, "db", "db.bolt", "")
}

func main() {
	flag.Parse()

	fmt.Println(host)
	fmt.Println(port)
	fmt.Println(dbFilename)

	ctx := context.Background()

	tr := &http.Transport{}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   1 * time.Second,
	}

	resizer := &domain.ContentTypeResizer{
		Mapping: map[string]domain.Resizer{
			"image/jpeg": &domain.JPEG{},
			"image/jpg":  &domain.JPEG{},
		},
		DefaultResizer: &domain.Proxy{},
	}
	domainInstance, err := domain.New(httpClient, resizer, dbFilename)
	if err != nil {
		log.Fatalf("Cannot run domain instance: %s", err)
	}

	logger := log.New(os.Stdout, "facade: ", log.LstdFlags)
	facadeInstance, err := facade.New(logger, host, port, domainInstance)
	if err != nil {
		log.Fatalf("Cannot run facade instance: %s", err)
	}

	facadeInstance.Run(ctx)

	select {}
}
