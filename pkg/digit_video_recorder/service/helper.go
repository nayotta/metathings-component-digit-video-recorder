package digit_video_recorder_service

import (
	"github.com/golang/protobuf/ptypes"

	driver "github.com/nayotta/metathings-component-digit-video-recorder/pkg/digit_video_recorder/driver"
	pb "github.com/nayotta/metathings-component-digit-video-recorder/proto"
)

func copy_record(x *driver.Record) *pb.Record {
	start_at, _ := ptypes.TimestampProto(x.StartAt)
	end_at, _ := ptypes.TimestampProto(x.EndAt)
	y := &pb.Record{
		Id:      x.Id,
		StartAt: start_at,
		EndAt:   end_at,
	}

	return y
}

func copy_records(xs []*driver.Record) []*pb.Record {
	var ys []*pb.Record
	for _, x := range xs {
		ys = append(ys, copy_record(x))
	}
	return ys
}
