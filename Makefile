# hex
# See LICENSE for copyright and license details.
.POSIX:

run:
	go run ./cmd/hex/

watch:
	fd -e go -e tmpl | entr -rs "go run ./cmd/hex/ -hb http://localhost:4001/"

faker:
	go run ./cmd/faker/

.PHONY: run watch faker
