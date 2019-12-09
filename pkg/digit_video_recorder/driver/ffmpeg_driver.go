package digit_video_recorder_driver

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"

	id_helper "github.com/nayotta/metathings/pkg/common/id"
	opt_helper "github.com/nayotta/metathings/pkg/common/option"
)

/*
 * Driver: ffmpeg
 *   record video by ffmpeg toolkit.
 * Options:
 *   driver:
 *     name: ffmpeg
 *     input:
 *       format: <format>  // input file format, like `v4l2`.
 *       file: <path>  // file path, like `/dev/video0` etc.
 *       [ frame_size: <width>x<height> ]  // frame size, like `640x480`.
 *       [ frame_rate: <rate> ]  // frame rate, like `30`.
 *     output:
 *       format: <format>  // output file format, like `mp4`.
 *       segment_time: <sec>  // segment time
 *       file: <path>  // video file, path template supported.
 *                     // fields:
 *                     //   id: video id, 32 bytes.
 *                     //   start_at: timestamp, recording start at the time
 *                     //   end_at: timestamp, recording end at the time
 *                     // example: /myvideo/{.id}-{.start_at}-{.end_at}.mp4
 *     video:
 *       codec:
 *         name: <codec>  // video codec, like `h264_omx` for raspberry pi.
 *         [ bit_rate: <rate> ]  // video bitrate, like `2000k`.
 *         [ extra: [ ... ] ]  // list of extra arguments for codec.
 *     audio:
 *       codec:
 *         name: <codec>  // audio codec, like `copy` for copy rtsp to file
 */

const (
	FFMPEG_DEFAULT_BINARY = `ffmpeg`
)

type FFmpegDigitVideoRecorderDriver struct {
	op_mtx            sync.Mutex
	cfn               context.CancelFunc
	tmp_dir           string
	watcher           *fsnotify.Watcher
	fs_evt_ch         chan fsnotify.Event
	cmd               *exec.Cmd
	logger            log.FieldLogger
	opt               *DigitVideoRecorderDriverOption
	st                *DigitVideoRecorderState
	tmpl              *template.Template
	storage           RecordStorage
	writing_file_chan chan string
}

func (drv *FFmpegDigitVideoRecorderDriver) get_logger() log.FieldLogger {
	return drv.logger
}

func (drv *FFmpegDigitVideoRecorderDriver) is_valid_file(name string) bool {
	base := filepath.Base(name)
	return strings.HasPrefix(base, "mtdvr-")
}

func (drv *FFmpegDigitVideoRecorderDriver) watch_file_loop() {
	var cur string
	ch := drv.writing_file_chan
_watch_file_loop:
	for {
		select {
		case name, ok := <-ch:
			if !ok {
				break _watch_file_loop
			}

			if cur == name || !drv.is_valid_file(name) {
				continue
			}

			if cur != "" {
				if err := drv.process_file(cur); err != nil {
					drv.get_logger().WithError(err).WithField("file", name).Warningf("failed to process file")
				}
			}
			cur = name
		}
	}
	drv.get_logger().Debugf("watch file loop exit")
}

func (drv *FFmpegDigitVideoRecorderDriver) parse_ffmpeg_command() (string, error) {
	var cmd_str string
	var err error

	if val := drv.opt.GetString("binary"); val != "" {
		cmd_str = val
	} else {
		cmd_str = FFMPEG_DEFAULT_BINARY
	}

	cmd_str += " -y"

	// INPUT
	input := drv.opt.Sub("input")
	if input == nil {
		return "", new_invalid_config_error("input")
	}

	if val := input.GetString("format"); val != "" {
		cmd_str += " -f " + val
	}

	if val := input.GetString("file"); val != "" {
		cmd_str += " -i \"" + val + "\""
	} else {
		return "", new_invalid_config_error("input.file")
	}

	if val := input.GetString("frame_size"); val != "" {
		cmd_str += " -s " + val
	}

	if val := input.GetString("frame_rate"); val != "" {
		cmd_str += " -r " + val
	}

	// VIDEO
	video := drv.opt.Sub("video")
	if video == nil {
		return "", new_invalid_config_error("video")
	}

	video_codec := video.Sub("codec")
	if video_codec == nil {
		return "", new_invalid_config_error("video.codec")
	}

	if val := video_codec.GetString("name"); val != "" {
		cmd_str += " -c:v " + val
	} else {
		return "", new_invalid_config_error("video.codec.name")
	}

	if val := video_codec.GetString("bit_rate"); val != "" {
		cmd_str += " -b:v " + val
	}

	if val := video_codec.GetStringSlice("extra"); val != nil {
		cmd_str += " " + strings.Join(val, " ")
	}

	// AUDIO
	audio := drv.opt.Sub("audio")
	if audio == nil {
		cmd_str += " -an"
	} else {
		audio_codec := audio.Sub("codec")
		if audio_codec == nil {
			return "", new_invalid_config_error("audio.codec")
		}

		if val := audio_codec.GetString("name"); val != "" {
			cmd_str += " -c:a " + val
		} else {
			return "", new_invalid_config_error("audio.codec.name")
		}

		if val := audio_codec.GetStringSlice("extra"); val != nil {
			cmd_str += " " + strings.Join(val, " ")
		}
	}

	// OUTPUT
	output := drv.opt.Sub("output")
	if output == nil {
		return "", new_invalid_config_error("output")
	}

	var segment_format string
	cmd_str += " -f segment"
	if segment_format = output.GetString("format"); segment_format != "" {
		cmd_str += " -segment_format " + segment_format
	} else {
		return "", new_invalid_config_error("output.format")
	}

	if val := output.GetString("segment_time"); val != "" {
		cmd_str += " -segment_time " + val
	} else {
		return "", new_invalid_config_error("output.segment_time")
	}

	drv.tmp_dir, err = ioutil.TempDir("", "mt_mdl_dvr")
	if err != nil {
		return "", err
	}

	cmd_str += " " + path.Join(drv.tmp_dir, fmt.Sprintf("mtdvr-%d-%%08d.%v", time.Now().Unix(), segment_format))

	return cmd_str, nil
}

func (drv *FFmpegDigitVideoRecorderDriver) Reset() error {
	drv.op_mtx.Lock()
	defer drv.op_mtx.Unlock()

	return drv.reset()
}

func (drv *FFmpegDigitVideoRecorderDriver) reset() error {
	var err error

	if drv.writing_file_chan != nil {
		close(drv.writing_file_chan)
		drv.writing_file_chan = nil
	}

	if drv.cfn != nil {
		drv.cfn()
		drv.cfn = nil
	}

	if drv.watcher != nil {
		if err = drv.watcher.Close(); err != nil {
			return err
		}
		drv.watcher = nil
	}

	drv.st = DIGITI_VIDEO_RECORDER_STATE_OFF

	return nil
}

func (drv *FFmpegDigitVideoRecorderDriver) get_output_file_template() *template.Template {
	if drv.tmpl == nil {
		drv.tmpl = template.Must(template.New("driver").Parse(drv.opt.GetString("output.file")))
	}
	return drv.tmpl
}

func (drv *FFmpegDigitVideoRecorderDriver) process_file(path string) error {
	var ts, idx int
	var err error
	var buf strings.Builder

	base := filepath.Base(path)
	ext := filepath.Ext(base)

	if _, err = fmt.Sscanf(base, "mtdvr-%d-%d"+ext, &ts, &idx); err != nil {
		return err
	}

	id := id_helper.NewId()
	segment_time := drv.opt.GetInt("output.segment_time")
	start_at := ts + idx*segment_time
	end_at := start_at + segment_time

	r := &Record{
		Id:      id,
		StartAt: time.Unix(int64(start_at), 0),
		EndAt:   time.Unix(int64(end_at), 0),
	}

	if err = drv.get_output_file_template().Execute(&buf, r.Data()); err != nil {
		return err
	}
	r.Path = buf.String()

	if err = os.Rename(path, r.Path); err != nil {
		return err
	}

	if err = drv.storage.SetRecord(r); err != nil {
		return err
	}

	return nil
}

func (drv *FFmpegDigitVideoRecorderDriver) Start() error {
	drv.op_mtx.Lock()
	defer drv.op_mtx.Unlock()

	if drv.cfn != nil {
		drv.get_logger().WithError(ErrNotStartable).Debugf("ffmpeg not startable")
		return ErrNotStartable
	}

	cmd_str, err := drv.parse_ffmpeg_command()
	if err != nil {
		drv.get_logger().WithError(err).Debugf("failed to parse ffmpeg command")
		return err
	}

	drv.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		drv.get_logger().WithError(err).Debugf("failed to new filesystem watcher")
		return err
	}

	err = drv.watcher.Add(drv.tmp_dir)
	if err != nil {
		drv.get_logger().WithError(err).Debugf("failed to watch filesystem")
		return err
	}

	// init write file channel
	if drv.writing_file_chan == nil {
		drv.writing_file_chan = make(chan string)
	}

	go drv.watch_file_loop()
	go func() {
		defer drv.Reset()
		watcher := drv.watcher
	_fsnotify_loop:
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					break _fsnotify_loop
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					drv.writing_file_chan <- event.Name
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					break _fsnotify_loop
				}
				drv.get_logger().WithError(err).Errorf("failed to watch video dir change")
			}
		}
		drv.get_logger().Debugf("filesystem watcher exited")
	}()

	ctx := context.TODO()
	ctx, drv.cfn = context.WithCancel(ctx)
	drv.cmd = exec.CommandContext(ctx, "/bin/bash", "-c", cmd_str)

	err = drv.cmd.Start()
	if err != nil {
		return err
	}
	drv.st = DIGITI_VIDEO_RECORDER_STATE_ON

	go func() {
		err := drv.cmd.Wait()

		drv.op_mtx.Lock()
		defer drv.op_mtx.Unlock()

		if drv.st == DIGITI_VIDEO_RECORDER_STATE_OFF {
			return
		}

		if err != nil {
			drv.logger.WithError(err).Warningf("ffmpeg exit with error")
		}

		drv.reset()
	}()

	return nil
}

func (drv *FFmpegDigitVideoRecorderDriver) Stop() error {
	return drv.Reset()
}

func (drv *FFmpegDigitVideoRecorderDriver) State() *DigitVideoRecorderState {
	drv.op_mtx.Lock()
	defer drv.op_mtx.Unlock()

	return drv.st
}

func (drv *FFmpegDigitVideoRecorderDriver) GetRecord(id string) (*Record, error) {
	drv.op_mtx.Lock()
	defer drv.op_mtx.Unlock()

	return drv.storage.GetRecord(id)
}

func (drv *FFmpegDigitVideoRecorderDriver) ListRecords(flt ListRecordsFitler) ([]*Record, error) {
	drv.op_mtx.Lock()
	defer drv.op_mtx.Unlock()

	return drv.storage.ListRecords(flt)
}

func NewFFmpegDigitVideoRecorderDriver(opt *DigitVideoRecorderDriverOption, args ...interface{}) (DigitVideoRecorderDriver, error) {
	var logger log.FieldLogger

	opt_helper.Setopt(opt_helper.SetoptConds{
		"logger": opt_helper.ToLogger(&logger),
	})(args...)

	stor_opt := &RecordStorageOption{opt.Sub("storage").Viper}
	stor, err := NewRecordStorage(stor_opt.GetString("name"), stor_opt, "logger", logger)
	if err != nil {
		return nil, err
	}

	drv := &FFmpegDigitVideoRecorderDriver{
		logger:  logger,
		opt:     opt,
		storage: stor,
		st:      DIGITI_VIDEO_RECORDER_STATE_OFF,
	}

	drv.logger.Debugf("new ffmpeg digit video recorder")

	return drv, nil
}

var register_ffmpeg_digit_video_recorder_driver_once sync.Once

func init() {
	register_ffmpeg_digit_video_recorder_driver_once.Do(func() {
		register_digit_video_recorder_driver_factory("ffmpeg", NewFFmpegDigitVideoRecorderDriver)
	})
}
