package digit_video_recorder_driver

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Record struct {
	Id      string    `yaml:"id"`
	StartAt time.Time `yaml:"start_at"`
	EndAt   time.Time `yaml:"end_at"`
	Path    string    `yaml:"path"`
}

func (r *Record) Reader() (io.Reader, error) {
	return os.Open(r.Path)
}

func (r *Record) Data() map[string]interface{} {
	return map[string]interface{}{
		"id":       r.Id,
		"start_at": r.StartAt.Unix(),
		"end_at":   r.EndAt.Unix(),
		"path":     r.Path,
	}
}

type DigitVideoRecorderDriverOption struct {
	*viper.Viper
}

func (o *DigitVideoRecorderDriverOption) Sub(key string) *DigitVideoRecorderDriverOption {
	sub := o.Viper.Sub(key)
	if sub == nil {
		return nil
	}
	return &DigitVideoRecorderDriverOption{sub}
}

type DigitVideoRecorderState struct {
	state string
}

func (s *DigitVideoRecorderState) String() string {
	return s.state
}

var (
	DIGITI_VIDEO_RECORDER_STATE_ON  = &DigitVideoRecorderState{state: "on"}
	DIGITI_VIDEO_RECORDER_STATE_OFF = &DigitVideoRecorderState{state: "off"}
)

type DigitVideoRecorderDriver interface {
	Start() error
	Stop() error
	State() *DigitVideoRecorderState
	GetRecord(id string) (*Record, error)
	ListRecords(ListRecordsFitler) ([]*Record, error)
}

type DigitVideoRecorderDriverFactory func(opt *DigitVideoRecorderDriverOption, args ...interface{}) (DigitVideoRecorderDriver, error)

var digit_video_recorder_driver_factories map[string]DigitVideoRecorderDriverFactory
var digit_video_recorder_driver_factories_once sync.Once

func register_digit_video_recorder_driver_factory(name string, fty DigitVideoRecorderDriverFactory) {
	digit_video_recorder_driver_factories_once.Do(func() {
		digit_video_recorder_driver_factories = make(map[string]DigitVideoRecorderDriverFactory)
	})
	digit_video_recorder_driver_factories[name] = fty
}

func NewDigitVideoRecorderDriver(name string, opt *DigitVideoRecorderDriverOption, args ...interface{}) (DigitVideoRecorderDriver, error) {
	fty, ok := digit_video_recorder_driver_factories[name]
	if !ok {
		return nil, ErrInvalidDigitVideoRecorderDriver
	}
	return fty(opt, args...)
}
