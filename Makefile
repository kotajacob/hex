# hex
# See LICENSE for copyright and license details.
.POSIX:

PREFIX ?= /usr
GO ?= go
GOFLAGS ?= -buildvcs=false
RM ?= rm -f

all: hex

hex:
	$(GO) build $(GOFLAGS) ./cmd/hex/

install: all
	mkdir -p $(DESTDIR)$(PREFIX)/bin
	cp -f hex $(DESTDIR)$(PREFIX)/bin
	chmod 755 $(DESTDIR)$(PREFIX)/bin/hex

uninstall:
	$(RM) $(DESTDIR)$(PREFIX)/bin/hex

clean:
	$(RM) hex

run:
	go run ./cmd/hex/

watch:
	fd -e go -e tmpl | entr -rs "go run ./cmd/hex/ -hb http://localhost:4001/"

faker:
	go run ./cmd/faker/

.PHONY: all install uninstall clean run watch faker
