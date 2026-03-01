BINARY  := mpv-wallpaper-tui
DESTDIR := $(HOME)/.local/bin

.PHONY: build install uninstall

build:
	go build -o $(BINARY) .

install: build
	install -Dm755 $(BINARY) $(DESTDIR)/$(BINARY)

uninstall:
	rm -f $(DESTDIR)/$(BINARY)
