package action

import "testing"

func TestParseLsblkDisksIncludesPhysicalAndRaidDevices(t *testing.T) {
	output := `{
		"blockdevices": [
			{
				"name": "sda",
				"path": "/dev/sda",
				"size": 2000398934016,
				"model": "ST2000DM008",
				"type": "disk",
				"rota": true,
				"children": [
					{
						"name": "md0",
						"path": "/dev/md0",
						"size": "2000394749952",
						"model": "",
						"type": "raid1",
						"rota": false
					}
				]
			},
			{
				"name": "nvme0n1",
				"path": "/dev/nvme0n1",
				"size": "1024209543168",
				"model": "SAMSUNG MZVL21T0HCLR-00B00",
				"type": "disk",
				"rota": "0"
			},
			{
				"name": "loop0",
				"path": "/dev/loop0",
				"size": "123456",
				"model": "",
				"type": "loop",
				"rota": false
			}
		]
	}`

	disks := parseLsblkDisks(output)
	if len(disks) != 3 {
		t.Fatalf("expected 3 disks, got %d", len(disks))
	}

	if disks[0].Name != "/dev/sda" || disks[0].Type != "hdd" {
		t.Fatalf("unexpected physical disk parse result: %+v", disks[0])
	}

	if disks[1].Name != "/dev/md0" || disks[1].Type != "raid1" || disks[1].Model != "Linux Software RAID" {
		t.Fatalf("unexpected raid disk parse result: %+v", disks[1])
	}

	if disks[2].Name != "/dev/nvme0n1" || disks[2].Type != "nvme" {
		t.Fatalf("unexpected nvme disk parse result: %+v", disks[2])
	}
}

func TestParseLsblkDisksReturnsNilOnInvalidJson(t *testing.T) {
	if disks := parseLsblkDisks(`{invalid`); disks != nil {
		t.Fatalf("expected nil disks on invalid json, got %+v", disks)
	}
}
