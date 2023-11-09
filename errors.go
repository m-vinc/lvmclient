package lvmclient

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

const (
	errMethodNotFound = `Method "GetAll" with signature "s" on interface "org.freedesktop.DBus.Properties" doesn't exist` + "\n"
)

var (
	ErrInvalidParams = errors.New("invalid params")
)

var (
	ErrLogicalVolumeNotFound     = errors.New("logical_volume: not found")
	ErrLogicalVolumeAlreadyExist = errors.New("logical_volume: already exist")
)

var (
	ErrVolumeGroupNotFound = errors.New("volume_group: not found")
)

// --

type JobError struct {
	Code    int32
	Message string
}

func (je *JobError) Error() string {
	return fmt.Sprintf("job complete with error. code:%d, message:%s", je.Code, je.Message)
}

var lvmErrorRegexp = regexp.MustCompile(`\('(.*)', 'Exit code ([0-9]*), stderr =   (.*)\\n'\)`)

func IsLvmError(err error) (*LvmError, bool) {
	if err == nil {
		return nil, false
	}

	je, ok := err.(*JobError)
	if !ok {
		return nil, false
	}

	if je.Code == -1 {
		matches := lvmErrorRegexp.FindAllStringSubmatch(je.Message, -1)
		if matches == nil {
			return nil, false
		}

		code := matches[0][2]
		description := matches[0][3]

		lvmCode, err := strconv.ParseInt(code, 10, 64)
		if err != nil {
			return nil, false
		}

		return &LvmError{
			Code:        int32(lvmCode),
			Description: description,
		}, true
	}
	return nil, false
}

var (
	lvmLogicalVolumeAlreadyExist = regexp.MustCompile(`^Logical Volume "(.*)" already exists in volume group "(.*)"`)
)

type LvmError struct {
	Code        int32
	Description string
}

func (lerr *LvmError) ToError() error {
	switch lerr.Code {
	case 5:
		if matches := lvmLogicalVolumeAlreadyExist.FindAllStringSubmatch(lerr.Description, -1); matches != nil {
			return ErrLogicalVolumeAlreadyExist
		}
		fallthrough
	default:
		return fmt.Errorf("%d - %s", lerr.Code, lerr.Description)
	}
}
