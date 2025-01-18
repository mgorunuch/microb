package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/mgorunuch/microb/app/core"
)

type DomainData struct {
	Domain string
}

var uniqueFlag = flag.Bool("unique", true, "Only output unique domains")

func extractDomain(ctx context.Context, url string) (*DomainData, error) {
	domain, err := core.ParseUrlHostName(ctx, url)
	if err != nil {
		return nil, err
	}
	return &DomainData{Domain: domain}, nil
}

func main() {
	core.Init()
	core.ProcessLines(core.SimpleConfig[*DomainData]{
		ThreadsCount: 1,
		KeyFunc:      core.ParseUrlHostName,
		RunFunc: func(ctx context.Context, url string) (*DomainData, error) {
			return extractDomain(ctx, url)
		},
		OutputFunc: func(data *DomainData) {
			fmt.Println(data.Domain)
		},
		Unique: *uniqueFlag,
	})
}
