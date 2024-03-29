// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: service.proto

package ai_metathings_component_service_digit_video_recorder

import (
	fmt "fmt"
	math "math"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/golang/protobuf/ptypes/empty"
	_ "github.com/golang/protobuf/ptypes/wrappers"
	_ "github.com/golang/protobuf/ptypes/timestamp"
	github_com_mwitkow_go_proto_validators "github.com/mwitkow/go-proto-validators"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

func (this *Record) Validate() error {
	if this.StartAt != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.StartAt); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("StartAt", err)
		}
	}
	if this.EndAt != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.EndAt); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("EndAt", err)
		}
	}
	return nil
}
func (this *OpRecord) Validate() error {
	if this.Id != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Id); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Id", err)
		}
	}
	if this.StartAt != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.StartAt); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("StartAt", err)
		}
	}
	if this.EndAt != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.EndAt); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("EndAt", err)
		}
	}
	return nil
}
func (this *GetRecordRequest) Validate() error {
	if this.Record != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Record); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Record", err)
		}
	}
	return nil
}
func (this *GetRecordResponse) Validate() error {
	if this.Record != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.Record); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("Record", err)
		}
	}
	return nil
}
func (this *ListRecordsRequest) Validate() error {
	if oneOfNester, ok := this.GetFilter().(*ListRecordsRequest_Range); ok {
		if oneOfNester.Range != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(oneOfNester.Range); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Range", err)
			}
		}
	}
	return nil
}
func (this *ListRecordsRequestRange_) Validate() error {
	if this.StartAt != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.StartAt); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("StartAt", err)
		}
	}
	if this.EndAt != nil {
		if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(this.EndAt); err != nil {
			return github_com_mwitkow_go_proto_validators.FieldError("EndAt", err)
		}
	}
	return nil
}
func (this *ListRecordsResponse) Validate() error {
	for _, item := range this.Records {
		if item != nil {
			if err := github_com_mwitkow_go_proto_validators.CallValidatorIfExists(item); err != nil {
				return github_com_mwitkow_go_proto_validators.FieldError("Records", err)
			}
		}
	}
	return nil
}
