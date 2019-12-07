package digit_video_recorder_service

import (
	"context"
	"strings"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	driver "github.com/nayotta/metathings-component-digit-video-recorder/pkg/digit_video_recorder/driver"
	pb "github.com/nayotta/metathings-component-digit-video-recorder/proto"
	component "github.com/nayotta/metathings/pkg/component"

	_ "net/http/pprof"
)

type DigitVideoRecorderService struct {
	module *component.Module
	drv    driver.DigitVideoRecorderDriver
}

func (s *DigitVideoRecorderService) logger() log.FieldLogger {
	return s.module.Logger()
}

func (s *DigitVideoRecorderService) update_state() error {
	if err := s.module.PutObject("state", strings.NewReader(s.drv.State().String())); err != nil {
		s.logger().WithError(err).Errorf("failed to set digit video recorder state")
		return err
	}

	return nil
}

func (s *DigitVideoRecorderService) reset() {
	s.drv.Stop()
	s.update_state()
}

func (s *DigitVideoRecorderService) HANDLE_GRPC_Start(ctx context.Context, in *any.Any) (*any.Any, error) {
	var err error
	req := &empty.Empty{}

	if err = ptypes.UnmarshalAny(in, req); err != nil {
		return nil, err
	}

	res, err := s.Start(ctx, req)
	if err != nil {
		return nil, err
	}

	out, err := ptypes.MarshalAny(res)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *DigitVideoRecorderService) Start(ctx context.Context, _ *empty.Empty) (*empty.Empty, error) {
	var err error

	if err = s.drv.Start(); err != nil {
		s.module.Logger().WithError(err).Errorf("failed to start recorder")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if err = s.update_state(); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	s.module.Logger().Debugf("recorder start")

	return &empty.Empty{}, nil
}

func (s *DigitVideoRecorderService) HANDLE_GRPC_Stop(ctx context.Context, in *any.Any) (*any.Any, error) {
	var err error
	req := &empty.Empty{}

	if err = ptypes.UnmarshalAny(in, req); err != nil {
		return nil, err
	}

	res, err := s.Stop(ctx, req)
	if err != nil {
		return nil, err
	}

	out, err := ptypes.MarshalAny(res)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *DigitVideoRecorderService) Stop(ctx context.Context, _ *empty.Empty) (*empty.Empty, error) {
	var err error

	if err = s.drv.Stop(); err != nil {
		s.module.Logger().WithError(err).Errorf("failed to stop recorder")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if err = s.update_state(); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	s.module.Logger().Debugf("recorder stop")

	return &empty.Empty{}, nil
}

func (s *DigitVideoRecorderService) HANDLE_GRPC_GetRecord(ctx context.Context, in *any.Any) (*any.Any, error) {
	var err error
	req := &pb.GetRecordRequest{}

	if err = ptypes.UnmarshalAny(in, req); err != nil {
		return nil, err
	}

	res, err := s.GetRecord(ctx, req)
	if err != nil {
		return nil, err
	}

	out, err := ptypes.MarshalAny(res)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *DigitVideoRecorderService) GetRecord(ctx context.Context, req *pb.GetRecordRequest) (*pb.GetRecordResponse, error) {
	id_str := req.GetRecord().GetId().GetValue()
	r, err := s.drv.GetRecord(id_str)
	if err != nil {
		s.module.Logger().WithError(err).Debugf("failed to get record")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	res := &pb.GetRecordResponse{
		Record: copy_record(r),
	}

	return res, nil
}

func (s *DigitVideoRecorderService) HANDLE_GRPC_ListRecords(ctx context.Context, in *any.Any) (*any.Any, error) {
	var err error
	req := &pb.ListRecordsRequest{}

	if err = ptypes.UnmarshalAny(in, req); err != nil {
		return nil, err
	}

	res, err := s.ListRecords(ctx, req)
	if err != nil {
		return nil, err
	}

	out, err := ptypes.MarshalAny(res)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *DigitVideoRecorderService) ListRecords(ctx context.Context, req *pb.ListRecordsRequest) (*pb.ListRecordsResponse, error) {
	var err error

	rng := req.GetRange()
	flt := driver.ListRecordsFitler{}
	if req_start_at := rng.GetStartAt(); req_start_at != nil {
		flt.Range.StartAt, err = ptypes.Timestamp(req_start_at)
		if err != nil {
			s.module.Logger().WithError(err).Debugf("failed to get range.start_at field")
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}
	}
	if req_end_at := rng.GetEndAt(); req_end_at != nil {
		flt.Range.EndAt, err = ptypes.Timestamp(req_end_at)
		if err != nil {
			s.module.Logger().WithError(err).Debugf("failed to get range.end_at field")
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}
	}

	rs, err := s.drv.ListRecords(flt)
	if err != nil {
		s.module.Logger().WithError(err).Debugf("failed to list records")
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	res := &pb.ListRecordsResponse{
		Records: copy_records(rs),
	}

	s.module.Logger().Debugf("list records")

	return res, nil
}

func (s *DigitVideoRecorderService) InitModuleService(m *component.Module) error {
	var err error

	s.module = m

	drv_opt := &driver.DigitVideoRecorderDriverOption{s.module.Kernel().Config().Sub("driver").Raw()}
	s.drv, err = driver.NewDigitVideoRecorderDriver(drv_opt.GetString("name"), drv_opt, "logger", s.logger(), "module", s.module)
	if err != nil {
		return err
	}

	s.logger().WithField("driver", drv_opt.GetString("name")).Debugf("init digit video recorder driver")
	s.reset()

	return nil
}
