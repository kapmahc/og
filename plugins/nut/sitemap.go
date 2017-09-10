package nut

import "github.com/ikeikeikeike/go-sitemap-generator/stm"

var _sitemap []stm.URL

// AddSitemapURL add sitemap url
func AddSitemapURL(args ...string) {
	for _, u := range args {
		_sitemap = append(_sitemap, stm.URL{"loc": u})
	}
}
