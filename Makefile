# hex
# See LICENSE for copyright and license details.
.POSIX:

PREFIX ?= /usr
GO ?= go
GOFLAGS ?= -buildvcs=false
RM ?= rm -f

all: hex

hex:
	$(GO) build $(GOFLAGS) .

install: all
	mkdir -p $(DESTDIR)$(PREFIX)/bin
	cp -f hex $(DESTDIR)$(PREFIX)/bin
	chmod 755 $(DESTDIR)$(PREFIX)/bin/hex

uninstall:
	$(RM) $(DESTDIR)$(PREFIX)/bin/hex

clean:
	$(RM) hex

run:
	go run -race .

watch:
	fd -e go -e tmpl | entr -rcs "go run -race ."

.PHONY: all hex install uninstall clean run watch
