package digit_video_recorder_driver

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidDigitVideoRecorderDriver = errors.New("invalid digit video recorder driver")
	ErrInvalidRecordStorage            = errors.New("invalid record storage")
	ErrNotStartable                    = errors.New("not startable")
	ErrNotFound                        = errors.New("record not found")
)

func new_invalid_config_error(key string) error {
	return errors.New(fmt.Sprintf("invalid config: %s", key))
}
