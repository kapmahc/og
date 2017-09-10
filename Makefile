dist=build
pkg=github.com/kapmahc/og/plugins/nut

VERSION=`git rev-parse --short HEAD`
BUILD_TIME=`date -R`
AUTHOR_NAME=`git config --get user.name`
AUTHOR_EMAIL=`git config --get user.email`
COPYRIGHT=`head -n 1 LICENSE`
USAGE=`sed -n '3p' README.md`

build: backend frontend 
	tar jcf dist.tar.bz2 $(dist)


backend:
	go build -ldflags "-s -w -X ${pkg}.Version=${VERSION} -X '${pkg}.BuildTime=${BUILD_TIME}' -X '${pkg}.AuthorName=${AUTHOR_NAME}' -X ${pkg}.AuthorEmail=${AUTHOR_EMAIL} -X '${pkg}.Copyright=${COPYRIGHT}' -X '${pkg}.Usage=${USAGE}'" -o ${dist}/og main.go
	-cp -r locales db templates $(dist)/

frontend:
	cd dashboard && npm run build
	-cp -r dashboard/dist $(dist)/public


clean:
	-rm -r $(dist) dist.tar.bz2 dashboard/dist


init:
	go get -u github.com/kardianos/govendor
	govendor sync
	cd dashboard && npm install
