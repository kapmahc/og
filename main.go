package main

import (
	"log"
	"os"

	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/kapmahc/og/plugins/erp"
	_ "github.com/kapmahc/og/plugins/forum"
	_ "github.com/kapmahc/og/plugins/mall"
	"github.com/kapmahc/og/plugins/nut"
	_ "github.com/kapmahc/og/plugins/ops/mail"
	_ "github.com/kapmahc/og/plugins/ops/vpn"
	_ "github.com/kapmahc/og/plugins/pos"
	_ "github.com/kapmahc/og/plugins/survey"
)

func main() {
	if err := nut.Main(os.Args...); err != nil {
		log.Fatal(err)
	}
}
