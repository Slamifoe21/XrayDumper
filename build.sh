#!/bin/bash
go build -ldflags="-s -w -X main.version=1.0.0 -X main.commit=$(git rev-parse --short HEAD 2>/dev/null || echo none)" -trimpath -o ./output/xd-linux
