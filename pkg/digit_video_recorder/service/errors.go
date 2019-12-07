package digit_video_recorder_service

import "errors"

var (
	ErrNotStartable = errors.New("not startable")
	ErrNotStopable  = errors.New("not stopable")
)
