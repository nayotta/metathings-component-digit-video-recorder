package main

import (
	"os"

	service "github.com/nayotta/metathings-component-digit-video-recorder/pkg/digit_video_recorder/service"
	component "github.com/nayotta/metathings/pkg/component"
)

func main() {
	mdl, err := component.NewModule(os.Args[0], new(service.DigitVideoRecorderService))
	if err != nil {
		panic(err)
	}
	err = mdl.Launch()
	if err != nil {
		panic(err)
	}
}
