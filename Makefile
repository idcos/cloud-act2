# Go parameters
GOBUILD=go build
GOCLEAN=go clean
GOTEST=go test
GOPACKAGES=$(shell go list ./... | grep -v /vendor/ | sed 's/^_//')
COMMIT=$(git rev-parse HEAD)
DATE=$(date +'%Y-%m-%dT%H:%M:%m+08:00')

# Binary name of cli
CLI_BINARY_NAME=cloud-act2

# Base path used to install cloudact2
DESTDIR=/usr/yunji/cloud-act2/bin

.PHONY: build
build: cloudact2


.PHONY: cloudact2
cloudact2: clean
	make -C src/idcos.io/cloud-act2/
	mkdir -p cmd/etc/acl
	mkdir -p cmd/bin
	cp src/idcos.io/cloud-act2/cmd/cloud-act2/cloud-act2-server cmd/bin/cloud-act2-server
	cp src/idcos.io/cloud-act2/cmd/act2ctl/act2ctl cmd/bin/act2ctl
	cp src/idcos.io/cloud-act2/cmd/salt-event/salt-event cmd/bin/salt-event
	cp -r src/idcos.io/cloud-act2/conf/cloud-{act2,act2-proxy}.yaml cmd/etc
	cp -r src/idcos.io/cloud-act2/conf/acl/{model.conf,policy.csv} cmd/etc/acl
	echo "====== if you want use start-stop-daemon, you should build in scripts/start-stop-daemon ======"


.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f cmd/bin/cloud-act2-server
	rm -f cmd/bin/act2ctl
	rm -rf cmd


.PHONY: run
run: build
	./cmd/cloud-act2/cloud-act2 web start


.PHONY: check
check: fmt lint validate-swagger

.PHONY: fmt
fmt: ## run go fmt
	@echo $@
	@which gofmt
	@test -z "$$(gofmt -s -l . 2>/dev/null | grep -Fv 'vendor/' | grep -v ".pb.go$$" | tee /dev/stderr)" || \
		(echo "please format Go code with 'gofmt -s -w'" && false)
	@test -z "$$(find . -path ./vendor -prune -o ! -name timestamp.proto ! -name duration.proto -name '*.proto' -type f -exec grep -Hn -e "^ " {} \; | tee /dev/stderr)" || \
		(echo "please indent proto files with tabs only" && false)
	@test -z "$$(find . -path ./vendor -prune -o -name '*.proto' -type f -exec grep -Hn "Meta meta = " {} \; | grep -v '(gogoproto.nullable) = false' | tee /dev/stderr)" || \
		(echo "meta fields in proto files must have option (gogoproto.nullable) = false" && false)

.PHONY: lint
lint: ## run revive
	@echo $@
	@revive -config revive.toml -exclude=./src/idcos.io/cloud-act2/vendor/... -formatter friendly . ./...

.PHONY: stat
stat:
	@cloc --exclude-dir=vendor . 


.PHONY: validate-swagger
validate-swagger: ## run swagger validate
	@echo $@ 
	@swagger validate src/idcos.io/cloud-act2/api/swagger.yml
