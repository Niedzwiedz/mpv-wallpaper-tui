BINARY       := mpv-wallpaper-tui
DESTDIR      := $(HOME)/.local/bin
SERVICE_NAME := mpv-wallpaper
SERVICE_DIR  := $(HOME)/.config/systemd/user

.PHONY: build install uninstall install-service uninstall-service install-autostart uninstall-autostart

build:
	go build -o $(BINARY) ./cmd/$(BINARY)/

install: build
	install -Dm755 $(BINARY) $(DESTDIR)/$(BINARY)

uninstall:
	rm -f $(DESTDIR)/$(BINARY)

install-service: install
	install -Dm644 configs/$(SERVICE_NAME).service $(SERVICE_DIR)/$(SERVICE_NAME).service
	systemctl --user daemon-reload
	systemctl --user enable --now $(SERVICE_NAME)

uninstall-service:
	systemctl --user disable --now $(SERVICE_NAME) 2>/dev/null || true
	rm -f $(SERVICE_DIR)/$(SERVICE_NAME).service
	systemctl --user daemon-reload

install-autostart: install
	@if [ -d /run/user/$$(id -u)/systemd ]; then \
		$(MAKE) install-service; \
	else \
		install -Dm644 $(SERVICE_NAME).desktop $(HOME)/.config/autostart/$(SERVICE_NAME).desktop; \
		echo "Installed XDG autostart entry (no systemd user session detected)"; \
	fi

uninstall-autostart:
	@if [ -d /run/user/$$(id -u)/systemd ]; then \
		$(MAKE) uninstall-service; \
	else \
		rm -f $(HOME)/.config/autostart/$(SERVICE_NAME).desktop; \
		echo "Removed XDG autostart entry"; \
	fi
