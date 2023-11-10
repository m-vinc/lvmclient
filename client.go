package lvmclient

import (
	"context"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
)

type Client interface {
	GetVolumeGroup(ctx context.Context, params *GetVolumeGroupParams) (*VolumeGroup, error)
	GetVolumeGroups(ctx context.Context) ([]*VolumeGroup, error)
	GetLogicalVolumes(ctx context.Context, params *GetVolumeGroupParams) ([]*LogicalVolume, error)

	GetLogicalVolume(ctx context.Context, params *GetLogicalVolumeParams) (*LogicalVolume, error)
	ToggleLogicalVolume(ctx context.Context, params *ToggleLogicalVolumeParams) (bool, error)

	CreateLogicalVolume(ctx context.Context, params *CreateLogicalVolumeParams) (*LogicalVolume, error)
	ResizeLogicalVolume(ctx context.Context, params *ResizeLogicalVolumeParams) (*LogicalVolume, error)
	RemoveLogicalVolume(ctx context.Context, params *GetLogicalVolumeParams) error
}

var _ Client = (*client)(nil)

type client struct {
	conn *dbus.Conn

	timeout time.Duration
}

func New() (*client, error) {
	c, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, err
	}

	return &client{
		conn:    c,
		timeout: 10 * time.Second,
	}, nil
}

// waitFor use a job path and call the Wait function then parse the return code and message or the result.
func (c *client) waitFor(ctx context.Context, jobPath dbus.ObjectPath, timeout int) (any, error) {
	obj := c.conn.Object("com.redhat.lvmdbus1", jobPath)
	err := obj.CallWithContext(ctx, "com.redhat.lvmdbus1.Job.Wait", 0, timeout).Err
	if err != nil {
		return nil, err
	}

	var jobMap map[string]any
	obj = c.conn.Object("com.redhat.lvmdbus1", jobPath)
	err = obj.CallWithContext(ctx, "org.freedesktop.DBus.Properties.GetAll", 0, "com.redhat.lvmdbus1.Job").Store(&jobMap)
	if err != nil {
		return nil, err
	}

	errMap, ok := jobMap["GetError"].([]any)
	if !ok {
		return nil, fmt.Errorf("unable to get error in job return object")
	}

	code, ok := errMap[0].(int32)
	if !ok {
		return nil, fmt.Errorf("unable to get error code in job return object")
	}

	message, ok := errMap[1].(string)
	if !ok {
		return nil, fmt.Errorf("unable to get error message in job return object")
	}

	if code != 0 {
		return nil, &JobError{Code: code, Message: message}
	}

	return jobMap["Result"], nil
}
