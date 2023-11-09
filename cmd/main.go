package main

import (
	"context"
	"log"
	"lvmclient"
)

func main() {
	ctx := context.Background()

	lvm, err := lvmclient.New()
	if err != nil {
		log.Fatal(err)
	}

	vgs, err := lvm.GetVolumeGroups(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, vg := range vgs {
		log.Println(vg.Name)
	}

	studentDisks, err := lvm.GetVolumeGroup(ctx, &lvmclient.GetVolumeGroupParams{Name: "students_disks"})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(studentDisks)

	lvs, err := lvm.GetLogicalVolumes(ctx, &lvmclient.GetVolumeGroupParams{
		Identifier: &studentDisks.Identifier,
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println(lvs)

	err = lvm.RemoveLogicalVolume(ctx, &lvmclient.GetLogicalVolumeParams{
		Name: "vm",
		GetVolumeGroupParams: &lvmclient.GetVolumeGroupParams{
			Name: "students_disks",
		},
	})

	vg, err := lvm.GetVolumeGroup(ctx, &lvmclient.GetVolumeGroupParams{Name: "students_disks"})
	if err != nil {
		log.Fatal(err)
	}

	fthGaYvD, err := lvm.GetLogicalVolume(ctx, &lvmclient.GetLogicalVolumeParams{
		GetVolumeGroupParams: &lvmclient.GetVolumeGroupParams{Identifier: &vg.Identifier},
		Name:                 "fthGaYvD",
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", fthGaYvD)

	vm, err := lvm.CreateLogicalVolume(ctx, &lvmclient.CreateLogicalVolumeParams{
		GetVolumeGroupParams: &lvmclient.GetVolumeGroupParams{Identifier: &vg.Identifier},
		Name:                 "vm",
		Size:                 5000000000,
	})

	log.Println(vm, err)
}
