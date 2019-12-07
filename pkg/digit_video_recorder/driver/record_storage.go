package digit_video_recorder_driver

import (
	"time"

	opt_helper "github.com/nayotta/metathings/pkg/common/option"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"gopkg.in/yaml.v2"
)

type RecordStorageOption struct {
	*viper.Viper
}

type ListRecordsFitler struct {
	Range struct {
		StartAt time.Time
		EndAt   time.Time
	}
}

type RecordStorage interface {
	ListRecords(ListRecordsFitler) ([]*Record, error)
	GetRecord(id string) (*Record, error)
	SetRecord(*Record) error
	UnsetRecord(id string) error
}

/*
 * Driver: leveldb
 *   leveldb record storage
 * Options:
 *   storage:
 *     name: leveldb
 *     file: <path>  // leveldb file path
 */

type leveldbRecordStorage struct {
	db     *leveldb.DB
	opt    *RecordStorageOption
	logger log.FieldLogger
}

func (s *leveldbRecordStorage) get_logger() log.FieldLogger {
	return s.logger
}

func (s *leveldbRecordStorage) ListRecords(flt ListRecordsFitler) ([]*Record, error) {
	var rs []*Record

	iter := s.db.NewIterator(util.BytesPrefix([]byte("record.")), nil)
	for iter.Next() {
		buf := iter.Value()
		var r Record
		if err := yaml.Unmarshal(buf, &r); err != nil {
			return nil, err
		}
		rs = append(rs, &r)
	}

	return rs, nil
}

func (s *leveldbRecordStorage) GetRecord(id string) (*Record, error) {
	buf, err := s.db.Get([]byte("record."+id), nil)
	if err == leveldb.ErrNotFound {
		return nil, ErrNotFound
	} else {
		return nil, err
	}

	var r Record
	if err = yaml.Unmarshal(buf, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

func (s *leveldbRecordStorage) SetRecord(r *Record) error {
	buf, err := yaml.Marshal(r)
	if err != nil {
		return err
	}

	if err = s.db.Put([]byte("record."+r.Id), buf, nil); err != nil {
		return err
	}

	return nil
}

func (s *leveldbRecordStorage) UnsetRecord(id string) error {
	if err := s.db.Delete([]byte("record."+id), nil); err != nil {
		return err
	}
	return nil
}

func NewRecordStorage(name string, opt *RecordStorageOption, args ...interface{}) (RecordStorage, error) {
	if name != "leveldb" {
		return nil, ErrInvalidRecordStorage
	}

	var logger log.FieldLogger
	var db *leveldb.DB
	var err error

	if err = opt_helper.Setopt(opt_helper.SetoptConds{
		"logger": opt_helper.ToLogger(&logger),
	})(args...); err != nil {
		return nil, err
	}

	if val := opt.GetString("file"); val != "" {
		if db, err = leveldb.OpenFile(val, nil); err != nil {
			return nil, err
		}
	} else {
		return nil, new_invalid_config_error("file")
	}

	stor := &leveldbRecordStorage{
		db:     db,
		opt:    opt,
		logger: logger,
	}

	return stor, nil
}
