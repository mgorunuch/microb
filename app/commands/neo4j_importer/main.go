package main

import (
	"context"
	"fmt"
	"github.com/mgorunuch/microb/app/core"
	"github.com/mgorunuch/microb/app/core/cache"
	"github.com/mgorunuch/microb/app/core/neo4j"
	"github.com/mgorunuch/microb/app/engine/alienvault_passivedns"
	"github.com/mgorunuch/microb/app/engine/binary_edge"
	"github.com/mgorunuch/microb/app/engine/certspotter"
	"github.com/mgorunuch/microb/app/engine/commoncrawl"
	"github.com/mgorunuch/microb/app/engine/crt_sh"
	"github.com/mgorunuch/microb/app/engine/google_custom_search"
	"strings"
	"sync"
	"time"
)

// Type aliases for better readability
type (
	AlienVault       = alienvault_passivedns.PassiveDnsResp
	BinaryEdge       = binary_edge.BinaryEdgeResponse
	CertspotterResp  = []certspotter.Issuance
	CommonCrawlResp  = []commoncrawl.CrawlData
	CrtShResp        = []crt_sh.CertData
	GoogleSearchResp = google_custom_search.GoogleCustomSearchResponse
)

func main() {
	ctx := context.Background()

	core.Init()
	defer core.Closer(neo4j.Init(ctx))()

	// Process all data sources
	processDataSources(ctx)
}

func processDataSources(ctx context.Context) {
	processors := []func(context.Context) error{
		processAlienVaultRecords,
		processBinaryEdgeRecords,
		processCertspotterRecords,
		processCommonCrawlRecords,
		processCrtShRecords,
		processGoogleSearchRecords,
		processWebArchiveRecords,
	}

	var wg sync.WaitGroup
	for _, processor := range processors {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := processor(ctx); err != nil {
				core.Logger.Errorf("failed to process data source: %v", err)
			}
		}()
	}
	wg.Wait()
}

func processAlienVaultRecords(ctx context.Context) error {
	return cache.ReadAllFileCacheFiles[AlienVault](cache.ReadAllFileCacheFilesOpts[AlienVault]{
		Dir: core.CommandAlienvaultPassivednsCacheDir,
		Process: func(rec cache.FileCacheRecord[AlienVault]) error {
			resp, err := rec.Read()
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			records := make([]neo4j.DnsRecord, len(resp.PassiveDns))
			for i, dns := range resp.PassiveDns {
				records[i] = neo4j.DnsRecord{
					Hostname:   dns.Hostname,
					Address:    dns.Address,
					RecordType: dns.RecordType,
					AssetType:  dns.AssetType,
					Timestamp:  rec.Ts,
				}
			}

			err = neo4j.InsertDNSRecords(ctx, neo4j.InsertDnsRecordOpts{
				Records:      records,
				RunKey:       rec.Key,
				RunTimestamp: rec.Ts,
				CommandName:  core.CommandAlienvaultPassivedns,
			})
			if err != nil {
				return fmt.Errorf("failed to insert DNS records: %w", err)
			}

			core.Logger.Debugf("[AlienVault/%s/%v] Successfully processed %d DNS records",
				rec.Key, rec.Ts.Unix(), len(records))
			return nil
		},
	})
}

func processBinaryEdgeRecords(ctx context.Context) error {
	return cache.ReadAllFileCacheFiles[BinaryEdge](cache.ReadAllFileCacheFilesOpts[BinaryEdge]{
		Dir: core.CommandBinaryEdgeCacheDir,
		Process: func(rec cache.FileCacheRecord[BinaryEdge]) error {
			resp, err := rec.Read()
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			records := make([]neo4j.DnsRecord, len(resp.Events))
			for i, dns := range resp.Events {
				records[i] = neo4j.DnsRecord{
					Hostname:  dns,
					Timestamp: rec.Ts,
				}
			}

			err = neo4j.InsertDNSRecords(ctx, neo4j.InsertDnsRecordOpts{
				Records:      records,
				RunKey:       rec.Key,
				RunTimestamp: rec.Ts,
				CommandName:  core.CommandBinaryEdge,
			})
			if err != nil {
				return fmt.Errorf("failed to insert DNS records: %w", err)
			}

			core.Logger.Debugf("[BinaryEdge/%s/%v] Successfully processed %d DNS records",
				rec.Key, rec.Ts.Unix(), len(records))
			return nil
		},
	})
}

func processCertspotterRecords(ctx context.Context) error {
	return cache.ReadAllFileCacheFiles[CertspotterResp](cache.ReadAllFileCacheFilesOpts[CertspotterResp]{
		Dir: core.CommandCertspotterCacheDir,
		Process: func(rec cache.FileCacheRecord[CertspotterResp]) error {
			resp, err := rec.Read()
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			certs := make([]neo4j.CertspotterCert, len(resp))
			for i, cert := range resp {
				certs[i] = neo4j.CertspotterCert{
					ID:           cert.ID,
					TbsSHA256:    cert.TbsSHA256,
					CertSHA256:   cert.CertSHA256,
					DNSNames:     cert.DNSNames,
					PubkeySHA256: cert.PubkeySHA256,
					NotBefore:    cert.NotBefore,
					NotAfter:     cert.NotAfter,
					Revoked:      cert.Revoked,
				}
			}

			err = neo4j.InsertCertspotterRecords(ctx, neo4j.InsertCertspotterOpts{
				Certificates: certs,
				RunKey:       rec.Key,
				RunTimestamp: rec.Ts,
				CommandName:  core.CommandCertspotter,
			})
			if err != nil {
				return fmt.Errorf("failed to insert certificate records: %w", err)
			}

			core.Logger.Debugf("[Certspotter/%s/%v] Successfully processed %d certificate records",
				rec.Key, rec.Ts.Unix(), len(certs))
			return nil
		},
	})
}

func processCommonCrawlRecords(ctx context.Context) error {
	return cache.ReadAllFileCacheFiles[CommonCrawlResp](cache.ReadAllFileCacheFilesOpts[CommonCrawlResp]{
		Dir: core.CommandCommonCrawlCacheDir,
		Process: func(rec cache.FileCacheRecord[CommonCrawlResp]) error {
			resp, err := rec.Read()
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			webpages := make([]neo4j.CommonCrawlWebpage, len(resp))
			for i, webpage := range resp {
				timestamp, err := time.Parse("20060102150405", webpage.Timestamp)
				if err != nil {
					return fmt.Errorf("failed to parse timestamp: %w", err)
				}

				webpages[i] = neo4j.CommonCrawlWebpage{
					Urlkey:       webpage.Urlkey,
					Timestamp:    timestamp,
					URL:          webpage.Url,
					Mime:         webpage.Mime,
					MimeDetected: webpage.MimeDetected,
					Status:       webpage.Status,
					Digest:       webpage.Digest,
					Length:       webpage.Length,
					Offset:       webpage.Offset,
					Filename:     webpage.Filename,
					Languages:    webpage.Languages,
					Encoding:     webpage.Encoding,
				}
			}

			err = neo4j.InsertCommonCrawlRecords(ctx, neo4j.InsertCommonCrawlOpts{
				Webpages:     webpages,
				RunKey:       rec.Key,
				RunTimestamp: rec.Ts,
				CommandName:  core.CommandCommonCrawl,
			})
			if err != nil {
				return fmt.Errorf("failed to insert CommonCrawl records: %w", err)
			}

			core.Logger.Debugf("[CommonCrawl/%s/%v] Successfully processed %d webpage records",
				rec.Key, rec.Ts.Unix(), len(webpages))
			return nil
		},
	})
}

func processCrtShRecords(ctx context.Context) error {
	return cache.ReadAllFileCacheFiles[CrtShResp](cache.ReadAllFileCacheFilesOpts[CrtShResp]{
		Dir: core.CommandCrtShCacheDir,
		Process: func(rec cache.FileCacheRecord[CrtShResp]) error {
			resp, err := rec.Read()
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			certs := make([]neo4j.CrtshCert, len(resp))
			for i, cert := range resp {
				entryTimestamp, err := time.Parse("2006-01-02T15:04:05.000", cert.EntryTimestamp)
				if err != nil {
					return fmt.Errorf("failed to parse entry timestamp: %w", err)
				}

				notBefore, err := time.Parse("2006-01-02T15:04:05", cert.NotBefore)
				if err != nil {
					return fmt.Errorf("failed to parse not_before: %w", err)
				}

				notAfter, err := time.Parse("2006-01-02T15:04:05", cert.NotAfter)
				if err != nil {
					return fmt.Errorf("failed to parse not_after: %w", err)
				}

				certs[i] = neo4j.CrtshCert{
					ID:             cert.ID,
					IssuerCAID:     cert.IssuerCaID,
					IssuerName:     cert.IssuerName,
					CommonName:     cert.CommonName,
					NameValue:      cert.NameValue,
					SerialNumber:   cert.SerialNumber,
					EntryTimestamp: entryTimestamp,
					NotBefore:      notBefore,
					NotAfter:       notAfter,
				}
			}

			err = neo4j.InsertCrtshRecords(ctx, neo4j.InsertCrtshOpts{
				Certificates: certs,
				RunKey:       rec.Key,
				RunTimestamp: rec.Ts,
				CommandName:  core.CommandCrtSh,
			})
			if err != nil {
				return fmt.Errorf("failed to insert crt.sh records: %w", err)
			}

			core.Logger.Debugf("[crt.sh/%s/%v] Successfully processed %d certificate records",
				rec.Key, rec.Ts.Unix(), len(certs))
			return nil
		},
	})
}

func processGoogleSearchRecords(ctx context.Context) error {
	return cache.ReadAllFileCacheFiles[GoogleSearchResp](cache.ReadAllFileCacheFilesOpts[GoogleSearchResp]{
		Dir: core.CommandGoogleSearchCacheDir,
		Process: func(rec cache.FileCacheRecord[GoogleSearchResp]) error {
			resp, err := rec.Read()
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			results := make([]neo4j.GoogleSearchResult, len(resp.Items))
			for i, item := range resp.Items {
				results[i] = neo4j.GoogleSearchResult{
					Title:       item.Title,
					Link:        item.Link,
					Snippet:     item.Snippet,
					DisplayLink: item.DisplayLink,
					Timestamp:   rec.Ts,
				}
			}

			err = neo4j.InsertGoogleSearchRecords(ctx, neo4j.InsertGoogleSearchOpts{
				Results:      results,
				RunKey:       rec.Key,
				RunTimestamp: rec.Ts,
				CommandName:  core.CommandGoogleSearch,
			})
			if err != nil {
				return fmt.Errorf("failed to insert Google search records: %w", err)
			}

			core.Logger.Debugf("[GoogleSearch/%s/%v] Successfully processed %d search results",
				strings.Trim(rec.Key, "\n"), rec.Ts.Unix(), len(results))
			return nil
		},
	})
}

func processWebArchiveRecords(ctx context.Context) error {
	return cache.ReadAllFileCacheFiles[[]string](cache.ReadAllFileCacheFilesOpts[[]string]{
		Dir: core.CommandWebArchiveCacheDir,
		Process: func(rec cache.FileCacheRecord[[]string]) error {
			urls, err := rec.Read()
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			webArchiveUrls := make([]neo4j.WebArchiveURL, 0, len(urls))
			for _, rawURL := range urls {
				parsedURL, err := neo4j.ParseWebArchiveURL(rawURL, rec.Ts)
				if err != nil {
					core.Logger.Warnf("[WebArchive/%s/%v] Failed to parse URL %s: %v",
						rec.Key, rec.Ts.Unix(), rawURL, err)
					continue
				}
				webArchiveUrls = append(webArchiveUrls, *parsedURL)
			}

			err = neo4j.InsertWebArchiveRecords(ctx, neo4j.InsertWebArchiveOpts{
				URLs:         webArchiveUrls,
				RunKey:       rec.Key,
				RunTimestamp: rec.Ts,
				CommandName:  core.CommandWebArchive,
			})
			if err != nil {
				return fmt.Errorf("failed to insert web archive records: %w", err)
			}

			core.Logger.Debugf("[WebArchive/%s/%v] Successfully processed %d URLs",
				rec.Key, rec.Ts.Unix(), len(webArchiveUrls))
			return nil
		},
	})
}
