run-ssh-server:
	make build-ssh-server
	bin/proxio -address=localhost -port=2222 -key="config/keys/server_id_rsa"
build-ssh-server:
	cd src/proxio && go build -o ../../bin/proxio ./cmd/server/server.go
ui:
	cd src/telemetry && npm install && ng build