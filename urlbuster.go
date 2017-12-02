// URL buster

package main

import (
	"fmt"

	uuid "github.com/satori/go.uuid"
)

func setupURL(cfg *config) bool {
	guid := uuid.NewV4()
	wildcardResp, _ := get(cfg, cfg.url, guid.String(), cfg.cookies)

	if cfg.statusCodes.contains(*wildcardResp) {
		cfg.isWildcard = true
		fmt.Println("[-] Wildcard response found:", fmt.Sprintf("%s%s", cfg.url, guid), "=>", *wildcardResp)
		if !cfg.wildcardForced {
			fmt.Println("[-] To force processing of Wildcard responses, specify the '-fw' switch.")
		}
		return cfg.wildcardForced
	}

	return true
}

func processURL(cfg *config, word string, brc chan<- busterResult) {
	suffix := ""
	if cfg.useSlash {
		suffix = "/"
	}

	// Try the DIR first
	dirResp, dirSize := get(cfg, cfg.url, word+suffix, cfg.cookies)
	if dirResp != nil {
		brc <- busterResult{
			entity: word + suffix,
			status: *dirResp,
			size:   dirSize,
		}
	}

	// Follow up with files using each ext.
	for ext := range cfg.extensions {
		file := word + cfg.extensions[ext]
		fileResp, fileSize := get(cfg, cfg.url, file, cfg.cookies)

		if fileResp != nil {
			brc <- busterResult{
				entity: file,
				status: *fileResp,
				size:   fileSize,
			}
		}
	}
}
