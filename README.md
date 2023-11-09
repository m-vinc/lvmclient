# LVMClient

This package aim to simplify interations with LVM while using Go. To do that we use [lvmdbusd](https://github.com/lvmteam/lvm2/tree/main/daemons) and the [dbus](https://github.com/godbus/dbus) package.

On ubuntu and debian you can install `lvmdbusd` :

```
$ apt install lvm2-dbusd
```

Then you're ready to go and use lvmclient :

```go
myVolumeGroup, err := lvm.GetVolumeGroup(ctx, &lvmclient.GetVolumeGroupParams{Name: "my_volume_group"})
if err != nil {
    return nil, err
}

log.Printf("%+v", myVolumeGroup)
```

Actually the client does not support many operations but I'll add them while working with LVM :

```go
type Client interface {
	GetVolumeGroup(ctx context.Context, params *GetVolumeGroupParams) (*VolumeGroup, error)
	GetVolumeGroups(ctx context.Context) ([]*VolumeGroup, error)
	GetLogicalVolumes(ctx context.Context, params *GetVolumeGroupParams) ([]*LogicalVolume, error)

	GetLogicalVolume(ctx context.Context, params *GetLogicalVolumeParams) (*LogicalVolume, error)
	ToggleLogicalVolume(ctx context.Context, params *GetLogicalVolumeParams, state bool) (bool, error)

	CreateLogicalVolume(ctx context.Context, params *CreateLogicalVolumeParams) (*LogicalVolume, error)
	RemoveLogicalVolume(ctx context.Context, params *GetLogicalVolumeParams) error
}
```

# Contributions

If you found any bugs or you want to implement a missing new functions, don't hesitate to do a PR.