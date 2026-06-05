.PHONY: all core console relay agent windows-agent web clean

all: core console relay agent

core:
	go build -o bin/ghost-core ./cmd/ghost-core

console:
	go build -o bin/ghost-console ./cmd/ghost-console

relay:
	go build -o bin/ghost-relay ./cmd/ghost-relay

agent:
	cd pkg/agent && cargo build --release
	cp pkg/agent/target/release/ghost-agent bin/

windows-agent:
	cd pkg/agent && cargo build --release --target x86_64-pc-windows-gnu
	cp pkg/agent/target/x86_64-pc-windows-gnu/release/ghost-agent.exe bin/

web:
	cd web-ui && npm install && npm run build
	cp -r web-ui/build bin/webui

clean:
	rm -rf bin/ ghost.db
