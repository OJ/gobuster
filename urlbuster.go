// URL buster

package main

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

func setupURL(cfg *config) bool {
	guid := uuid.NewV4()
	wildcardResp, _ := get(cfg, cfg.Url, guid.String(), cfg.Cookies)

	if cfg.StatusCodes.contains(*wildcardResp) {
		cfg.IsWildcard = true
		fmt.Println("[-] Wildcard response found:", fmt.Sprintf("%s%s", cfg.Url, guid), "=>", *wildcardResp)
		if !cfg.WildcardForced {
			fmt.Println("[-] To force processing of Wildcard responses, specify the '-fw' switch.")
		}
		return cfg.WildcardForced
	}

	return true
}

func processURL(cfg *config, word string, brc chan<- busterResult) {
	suffix := ""
	if cfg.UseSlash {
		suffix = "/"
	}

	// Try the DIR first
	dirResp, dirSize := get(cfg, cfg.Url, word+suffix, cfg.Cookies)
	if dirResp != nil {
		brc <- busterResult{
			Entity: word + suffix,
			Status: *dirResp,
			Size:   dirSize,
		}
	}

	// Follow up with files using each ext.
	for ext := range cfg.Extensions {
		file := word + cfg.Extensions[ext]
		fileResp, fileSize := get(cfg, cfg.Url, file, cfg.Cookies)

		if fileResp != nil {
			brc <- busterResult{
				Entity: file,
				Status: *fileResp,
				Size:   fileSize,
			}
		}
	}
}
