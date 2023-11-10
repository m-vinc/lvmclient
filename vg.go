package lvmclient

import (
	"context"
	"fmt"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/mitchellh/mapstructure"
)

// GetVolumeGroupParams is used to get a volume group either by Identifier or Name
type GetVolumeGroupParams struct {
	Identifier *Identifier
	Name       string
}

type VolumeGroup struct {
	Identifier Identifier `mapstructure:"-"`

	Name            string
	Tags            []string
	Pvs             []Identifier
	Lvs             []Identifier
	Writeable       bool
	Readable        bool
	Resizeable      bool
	Exportable      bool
	Partial         bool
	AllocContiguous bool
	AllocCling      bool
	AllocNormal     bool
	AllocAnywhere   bool
	Clustered       bool
	MdaUsedCount    uint64
	MdaSizeBytes    uint64
	MdaFree         uint64
	MdaCount        uint64
	Seqno           uint64
	SnapCount       uint64
	LvCount         uint64
	PvCount         uint64
	MaxPv           uint64
	MaxLv           uint64
	Profile         string
	FreeCount       uint64
	ExtentCount     uint64
	ExtentSizeBytes uint64
	SysId           string
	FreeBytes       uint64
	SizeBytes       uint64
	Fmt             string
	Uuid            string
}

// GetVolumeGroup get a volume group by Name or Identifier using a GetVolumeGroupParams
func (c *client) GetVolumeGroup(ctx context.Context, params *GetVolumeGroupParams) (*VolumeGroup, error) {
	var res string

	if params.Identifier != nil {
		vgMap := map[string]any{}
		vg := &VolumeGroup{}

		objPath := dbus.ObjectPath(*params.Identifier)
		obj := c.conn.Object("com.redhat.lvmdbus1", objPath)
		err := obj.CallWithContext(ctx, "org.freedesktop.DBus.Properties.GetAll", 0, "com.redhat.lvmdbus1.Vg").Store(&vgMap)
		notfound := err != nil && err.Error() == errMethodNotFound
		if err != nil && !notfound {
			return nil, err
		}

		if notfound {
			return nil, ErrVolumeGroupNotFound
		}

		err = mapstructure.Decode(vgMap, vg)
		if err != nil {
			return nil, err
		}
		vg.Identifier = Identifier(objPath)
		return vg, nil
	} else if params.Name != "" {
		obj := c.conn.Object("com.redhat.lvmdbus1", "/com/redhat/lvmdbus1/Vg")
		err := obj.CallWithContext(ctx, "org.freedesktop.DBus.Introspectable.Introspect", 0).Store(&res)
		if err != nil {
			return nil, err
		}

		node := &introspect.Node{}
		if err := customUnmarshal[introspect.Node]([]byte(res), node); err != nil {
			return nil, err
		}

		for _, childNode := range node.Children {
			vg := &VolumeGroup{}
			vgMap := map[string]any{}

			objPath := dbus.ObjectPath("/com/redhat/lvmdbus1/Vg/" + childNode.Name)
			obj := c.conn.Object("com.redhat.lvmdbus1", objPath)
			err := obj.CallWithContext(ctx, "org.freedesktop.DBus.Properties.GetAll", 0, "com.redhat.lvmdbus1.Vg").Store(&vgMap)
			notfound := err != nil && err.Error() == errMethodNotFound
			if err != nil && !notfound {
				return nil, err
			}

			if notfound {
				continue
			}

			if vgMap["Name"] == params.Name {
				err = mapstructure.Decode(vgMap, vg)
				if err != nil {
					return nil, err
				}
				vg.Identifier = Identifier(objPath)
				return vg, nil
			}
		}
		return nil, fmt.Errorf("volume group not found")
	}

	return nil, fmt.Errorf("invalid params")
}

// GetVolumeGroups list all the volume groups
func (c *client) GetVolumeGroups(ctx context.Context) ([]*VolumeGroup, error) {
	var res string

	obj := c.conn.Object("com.redhat.lvmdbus1", "/com/redhat/lvmdbus1/Vg")
	err := obj.CallWithContext(ctx, "org.freedesktop.DBus.Introspectable.Introspect", 0).Store(&res)
	if err != nil {
		return nil, err
	}

	node := &introspect.Node{}
	if err := customUnmarshal[introspect.Node]([]byte(res), node); err != nil {
		return nil, err
	}

	vgs := make([]*VolumeGroup, len(node.Children))
	for _, childNode := range node.Children {
		vg := &VolumeGroup{}
		vgMap := map[string]any{}

		objPath := dbus.ObjectPath("/com/redhat/lvmdbus1/Vg/" + childNode.Name)
		obj := c.conn.Object("com.redhat.lvmdbus1", objPath)
		err := obj.CallWithContext(ctx, "org.freedesktop.DBus.Properties.GetAll", 0, "com.redhat.lvmdbus1.Vg").Store(&vgMap)
		notfound := err != nil && err.Error() == errMethodNotFound
		if err != nil && !notfound {
			return nil, err
		}

		if notfound {
			continue
		}

		err = mapstructure.Decode(vgMap, vg)
		if err != nil {
			return nil, err
		}
		vg.Identifier = Identifier(objPath)
		vgs = append(vgs, vg)
	}

	return vgs, nil
}
