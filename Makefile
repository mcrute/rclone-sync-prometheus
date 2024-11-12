VAULT_MATERIAL ?= some/example/material
B2_BUCKET ?= your-b2-bucket
INSTANCE_NAME ?= your.instance.example.com
PUSH_GATEWAY ?= https://pushgateway.example.com

rclone-report-arm64: *.go
	CGO_ENABLED=0 GOARCH=arm64 go build \
		-ldflags " \
			-X main.defaultVaultMaterial=$(VAULT_MATERIAL) \
			-X main.defaultB2Bucket=$(B2_BUCKET) \
			-X main.defaultInstanceName=$(INSTANCE_NAME) \
			-X main.defaultPushGateway=$(PUSH_GATEWAY) \
		"  \
		-o $@ $^

rclone-report-amd64: *.go
	CGO_ENABLED=0 go build \
		-ldflags " \
			-X main.defaultVaultMaterial=$(VAULT_MATERIAL) \
			-X main.defaultB2Bucket=$(B2_BUCKET) \
			-X main.defaultInstanceName=$(INSTANCE_NAME) \
			-X main.defaultPushGateway=$(PUSH_GATEWAY) \
		"  \
		-o $@ $^

.PHONY: clean
clean:
	rm rclone-report-arm64 || true
	rm rclone-report-amd64 || true
