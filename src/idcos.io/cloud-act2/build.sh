#!/bin/bash

COMMIT=$(git rev-parse HEAD) 
DATE=$(date +'%Y-%m-%dT%H:%M:%m+08:00') 
BRANCH=$(git branch | grep \* | cut -d ' ' -f2)

go build -o cmd/cloud-act2/cloud-act2-server -ldflags "-X \"idcos.io/cloud-act2/build.Date=$DATE\" -X \"idcos.io/cloud-act2/build.Commit=$COMMIT\" -X \"idcos.io/cloud-act2/build.GitBranch=$BRANCH\"" cmd/cloud-act2/main.go
go build -o cmd/act2ctl/act2ctl -ldflags "-X \"idcos.io/cloud-act2/build.Date=$DATE\" -X \"idcos.io/cloud-act2/build.Commit=$COMMIT\" -X \"idcos.io/cloud-act2/build.GitBranch=$BRANCH\"" cmd/act2ctl/*.go

