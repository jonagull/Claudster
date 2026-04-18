BIN=claudster
INSTALL_PATH=/usr/local/bin/$(BIN)

build:
	go build -o $(BIN) .

install: build
	cp $(BIN) $(INSTALL_PATH)
	codesign --sign - $(INSTALL_PATH)
	@echo "installed to $(INSTALL_PATH)"

uninstall:
	rm -f $(INSTALL_PATH)
	@echo "removed $(INSTALL_PATH)"

.PHONY: build install uninstall
