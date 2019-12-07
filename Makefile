DOCKER_BIN=$(shell which docker)

all:
	go build -o metathings-module-digit-video-recorder cmd/digit_video_recorder/main.go

protos:
	$(DOCKER_BIN) run --rm -v $(PWD):/go/src/github.com/nayotta/metathings-component-digit-video-recorder nayotta/metathings-development /usr/bin/make -C /go/src/github.com/nayotta/metathings-component-digit-video-recorder/proto
