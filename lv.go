package lvmclient

import (
	"context"
	"fmt"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/mitchellh/mapstructure"
)

// Identifier is a custom type we're using to not expose any dbus things
type Identifier dbus.ObjectPath

// GetLogicalVolumeParams is used to retrieve a volume group. You need to either specify a Identifier or a GetVolumeGroupParams and a Name
type GetLogicalVolumeParams struct {
	GetVolumeGroupParams *GetVolumeGroupParams
	Name                 string

	Identifier *Identifier
}

// CreateLogicalVolumeParams is used to retrieve a create a new Logical Volume, you need to specify in which volume group with the GetVolumeGroupParams.
// Size must be in bytes
type CreateLogicalVolumeParams struct {
	GetVolumeGroupParams *GetVolumeGroupParams

	Name string
	Size uint64
}

type LogicalVolume struct {
	Identifier Identifier `mapstructure:"-"`

	VolumeType        []string
	Permissions       []string
	AllocationPolicy  []string
	FixedMinor        bool
	State             []string
	TargetType        []string
	ZeroBlocks        bool
	Health            []string
	SkipActivation    bool
	Tags              []string
	Roles             []string
	IsThinVolume      bool
	IsThinPool        bool
	Active            bool
	MovePv            string
	MetaDataSizeBytes uint64
	SyncPercent       uint32
	CopyPercent       uint32
	MetaDataPercent   uint32
	SnapPercent       uint32
	DataPercent       uint32
	Attr              string
	HiddenLvs         []string
	Devices           []any
	PoolLv            string
	OriginLv          string
	Vg                string
	SegType           []string
	SizeBytes         uint64
	Path              string
	Name              string
	Uuid              string
}

// GetLogicalVolumes get all the logical volumes of a volume group
func (c *client) GetLogicalVolumes(ctx context.Context, params *GetVolumeGroupParams) ([]*LogicalVolume, error) {
	vg, err := c.GetVolumeGroup(ctx, params)
	if err != nil {
		return nil, err
	}

	lvs := make([]*LogicalVolume, len(vg.Lvs))
	for i, lvPath := range vg.Lvs {
		lv := &LogicalVolume{}
		lvMap := map[string]any{}

		obj := c.conn.Object("com.redhat.lvmdbus1", dbus.ObjectPath(lvPath))
		err := obj.CallWithContext(ctx, "org.freedesktop.DBus.Properties.GetAll", 0, "com.redhat.lvmdbus1.LvCommon").Store(&lvMap)
		if err != nil {
			return nil, err
		}

		err = mapstructure.Decode(lvMap, lv)
		if err != nil {
			return nil, err
		}

		lvs[i] = lv
	}

	return lvs, nil
}

// GetLogicalVolume get a logical volume by name or by identifier
func (c *client) GetLogicalVolume(ctx context.Context, params *GetLogicalVolumeParams) (*LogicalVolume, error) {
	if params.Identifier != nil {
		lv := &LogicalVolume{}
		lvMap := map[string]any{}

		lvPath := dbus.ObjectPath(*params.Identifier)
		obj := c.conn.Object("com.redhat.lvmdbus1", lvPath)
		err := obj.CallWithContext(ctx, "org.freedesktop.DBus.Properties.GetAll", 0, "com.redhat.lvmdbus1.LvCommon").Store(&lvMap)
		if err != nil {
			return nil, err
		}

		err = mapstructure.Decode(lvMap, lv)
		if err != nil {
			return nil, err
		}

		lv.Identifier = Identifier(lvPath)
		return lv, nil
	} else if params.Name != "" && params.GetVolumeGroupParams != nil {
		vg, err := c.GetVolumeGroup(ctx, params.GetVolumeGroupParams)
		if err != nil {
			return nil, err
		}

		for _, lvPath := range vg.Lvs {
			lv := &LogicalVolume{}
			lvMap := map[string]any{}

			obj := c.conn.Object("com.redhat.lvmdbus1", dbus.ObjectPath(lvPath))
			err := obj.CallWithContext(ctx, "org.freedesktop.DBus.Properties.GetAll", 0, "com.redhat.lvmdbus1.LvCommon").Store(&lvMap)
			if err != nil {
				return nil, err
			}

			if lvMap["Name"] == params.Name {
				err = mapstructure.Decode(lvMap, lv)
				if err != nil {
					return nil, err
				}

				lv.Identifier = lvPath
				return lv, nil
			}
		}

		return nil, fmt.Errorf("logical volume not found")
	}
	return nil, fmt.Errorf("invalid params, must include Identifier or Name")
}

// ToggleLogicalVolume Enable or disable a logical volume. use state to specify the desired state
func (c *client) ToggleLogicalVolume(ctx context.Context, params *GetLogicalVolumeParams, state bool) (bool, error) {
	lv, err := c.GetLogicalVolume(ctx, params)
	if err != nil {
		return false, err
	}

	method := "Deactivate"
	if state {
		method = "Activate"
	}

	var jobPath dbus.ObjectPath
	obj := c.conn.Object("com.redhat.lvmdbus1", dbus.ObjectPath(lv.Identifier))
	err = obj.CallWithContext(ctx, "com.redhat.lvmdbus1.Lv."+method, 0, 0, 0, map[string]any{}).Store(&jobPath)
	if err != nil {
		return false, err
	}

	_, err = c.waitFor(ctx, jobPath, int(c.timeout/time.Second))
	if err != nil {
		return false, err
	}

	return state, nil
}

// CreateLogicalVolume create a new logical volume into a specific volume group selected by the GetVolumeGroupParams
func (c *client) CreateLogicalVolume(ctx context.Context, params *CreateLogicalVolumeParams) (*LogicalVolume, error) {
	vg, err := c.GetVolumeGroup(ctx, params.GetVolumeGroupParams)
	if err != nil {
		return nil, err
	}

	var res []any
	obj := c.conn.Object("com.redhat.lvmdbus1", dbus.ObjectPath(vg.Identifier))
	err = obj.CallWithContext(ctx, "com.redhat.lvmdbus1.Vg.LvCreate", 0, params.Name, params.Size, map[string]any{}, 0, map[string]any{}).Store(&res)
	if err != nil {
		return nil, err
	}

	jobPath := res[1]
	result, err := c.waitFor(ctx, jobPath.(dbus.ObjectPath), int(c.timeout/time.Second))
	if err != nil {
		return nil, err
	}

	lvPath, ok := result.(dbus.ObjectPath)
	if !ok {
		return nil, fmt.Errorf("unable to get lvPath after end of creation")
	}

	identifierLvPath := Identifier(lvPath)

	return c.GetLogicalVolume(ctx, &GetLogicalVolumeParams{Identifier: &identifierLvPath})
}

// RemoveLogicalVolume remove a logical volume either by Name or Identifier
func (c *client) RemoveLogicalVolume(ctx context.Context, params *GetLogicalVolumeParams) error {
	lv, err := c.GetLogicalVolume(ctx, params)
	if err != nil {
		return err
	}

	var jobPath dbus.ObjectPath
	obj := c.conn.Object("com.redhat.lvmdbus1", dbus.ObjectPath(lv.Identifier))
	err = obj.CallWithContext(ctx, "com.redhat.lvmdbus1.Lv.Remove", 0, 0, map[string]any{}).Store(&jobPath)
	if err != nil {
		return err
	}

	_, err = c.waitFor(ctx, jobPath, int(c.timeout/time.Second))
	if err != nil {
		return err
	}

	return nil
}
