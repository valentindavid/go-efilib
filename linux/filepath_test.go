// Copyright 2021 Canonical Ltd.
// Licensed under the LGPLv3 with static-linking exception.
// See LICENCE file for details.

package linux_test

import (
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/canonical/go-efilib"
	. "github.com/canonical/go-efilib/linux"
)

type filepathSuite struct {
	FilepathMockMixin
	TarFileMixin
}

func (s *filepathSuite) mockOsOpen(m map[string]string) (restore func()) {
	return MockOsOpen(func(path string) (*os.File, error) {
		if p, ok := m[path]; ok {
			path = p
		}
		return os.Open(path)
	})
}

var _ = Suite(&filepathSuite{})

func (s *filepathSuite) TestFilePathToDevicePathShortFormFile(c *C) {
	restore := s.MockFilepathEvalSymlinks(map[string]string{})
	defer restore()
	restore = MockMountsPath("testdata/mounts-nvme")
	defer restore()
	restore = MockSysfsPath(filepath.Join(s.UnpackTar(c, "testdata/sys.tar"), "sys"))
	defer restore()

	path, err := FilePathToDevicePath("/boot/efi/EFI/ubuntu/shimx64.efi", ShortFormPathFile)
	c.Check(err, IsNil)
	c.Check(path, DeepEquals, efi.DevicePath{efi.FilePathDevicePathNode("\\EFI\\ubuntu\\shimx64.efi")})
}

func (s *filepathSuite) TestFilePathToDevicePathShortFormHD(c *C) {
	restore := s.MockFilepathEvalSymlinks(map[string]string{})
	defer restore()
	restore = MockMountsPath("testdata/mounts-nvme")
	defer restore()
	restore = s.mockOsOpen(map[string]string{"/dev/nvme0n1": "testdata/disk.img"})
	defer restore()
	restore = MockSysfsPath(filepath.Join(s.UnpackTar(c, "testdata/sys.tar"), "sys"))
	defer restore()

	path, err := FilePathToDevicePath("/boot/efi/EFI/ubuntu/shimx64.efi", ShortFormPathHD)
	c.Check(err, IsNil)
	c.Check(path, DeepEquals, efi.DevicePath{
		&efi.HardDriveDevicePathNode{
			PartitionNumber: 1,
			PartitionStart:  34,
			PartitionSize:   64,
			Signature:       efi.GUIDHardDriveSignature(efi.MakeGUID(0xc7a2907e, 0xd8c9, 0x4a41, 0x8b99, [...]uint8{0x3e, 0xf3, 0x24, 0x5f, 0xaf, 0x2a})),
			MBRType:         efi.GPT},
		efi.FilePathDevicePathNode("\\EFI\\ubuntu\\shimx64.efi")})
}

func (s *filepathSuite) TestFilePathToDevicePathShortFormHDUnpartitioned(c *C) {
	restore := s.MockFilepathEvalSymlinks(map[string]string{})
	defer restore()
	restore = MockMountsPath("testdata/mounts-nvme")
	defer restore()
	restore = MockSysfsPath(filepath.Join(s.UnpackTar(c, "testdata/sys.tar"), "sys"))
	defer restore()

	_, err := FilePathToDevicePath("/snap/core/11993/bin/ls", ShortFormPathHD)
	c.Check(err, ErrorMatches, "cannot map file path to a UEFI device path: file is not inside partitioned media - use linux.ShortFormPathFile")
}

func (s *filepathSuite) TestFilePathToDevicePathFullNVME(c *C) {
	restore := s.MockFilepathEvalSymlinks(map[string]string{})
	defer restore()
	restore = MockMountsPath("testdata/mounts-nvme")
	defer restore()
	restore = s.mockOsOpen(map[string]string{"/dev/nvme0n1": "testdata/disk.img"})
	defer restore()
	restore = MockSysfsPath(filepath.Join(s.UnpackTar(c, "testdata/sys.tar"), "sys"))
	defer restore()

	path, err := FilePathToDevicePath("/boot/efi/EFI/ubuntu/shimx64.efi", FullPath)
	c.Check(err, IsNil)
	c.Check(path, DeepEquals, efi.DevicePath{
		&efi.ACPIDevicePathNode{HID: 0x0a0341d0},
		&efi.PCIDevicePathNode{Function: 0, Device: 0x1d},
		&efi.PCIDevicePathNode{Function: 0, Device: 0x0},
		&efi.NVMENamespaceDevicePathNode{NamespaceID: 1},
		&efi.HardDriveDevicePathNode{
			PartitionNumber: 1,
			PartitionStart:  34,
			PartitionSize:   64,
			Signature:       efi.GUIDHardDriveSignature(efi.MakeGUID(0xc7a2907e, 0xd8c9, 0x4a41, 0x8b99, [...]uint8{0x3e, 0xf3, 0x24, 0x5f, 0xaf, 0x2a})),
			MBRType:         efi.GPT},
		efi.FilePathDevicePathNode("\\EFI\\ubuntu\\shimx64.efi")})
}

func (s *filepathSuite) TestFilePathToDevicePathFullSATA(c *C) {
	restore := s.MockFilepathEvalSymlinks(map[string]string{})
	defer restore()
	restore = MockMountsPath("testdata/mounts-sata")
	defer restore()
	restore = s.mockOsOpen(map[string]string{"/dev/sda": "testdata/disk.img"})
	defer restore()
	restore = MockSysfsPath(filepath.Join(s.UnpackTar(c, "testdata/sys.tar"), "sys"))
	defer restore()

	path, err := FilePathToDevicePath("/boot/efi/EFI/ubuntu/shimx64.efi", FullPath)
	c.Check(err, IsNil)
	c.Check(path, DeepEquals, efi.DevicePath{
		&efi.ACPIDevicePathNode{HID: 0x0a0341d0},
		&efi.PCIDevicePathNode{Function: 2, Device: 0x1f},
		&efi.SATADevicePathNode{
			HBAPortNumber:            0,
			PortMultiplierPortNumber: 0xffff},
		&efi.HardDriveDevicePathNode{
			PartitionNumber: 1,
			PartitionStart:  34,
			PartitionSize:   64,
			Signature:       efi.GUIDHardDriveSignature(efi.MakeGUID(0xc7a2907e, 0xd8c9, 0x4a41, 0x8b99, [...]uint8{0x3e, 0xf3, 0x24, 0x5f, 0xaf, 0x2a})),
			MBRType:         efi.GPT},
		efi.FilePathDevicePathNode("\\EFI\\ubuntu\\shimx64.efi")})
}

func (s *filepathSuite) TestFilePathToDevicePathFullNoDevicePath(c *C) {
	restore := s.MockFilepathEvalSymlinks(map[string]string{})
	defer restore()
	restore = MockMountsPath("testdata/mounts-nvme")
	defer restore()

	sysfs := filepath.Join(s.UnpackTar(c, "testdata/sys.tar"), "sys")
	restore = MockSysfsPath(sysfs)
	defer restore()

	_, err := FilePathToDevicePath("/snap/core/11993/bin/ls", FullPath)
	c.Check(err, ErrorMatches, "cannot map file path to a UEFI device path: encountered an error when handling components virtual/block/loop1 from device path "+sysfs+"/devices/virtual/block/loop1: \\[handler virtual\\]: unsupported device: virtual devices are not supported")
}

func (s *filepathSuite) TestFilePathToDevicePathFullWithBindMount(c *C) {
	restore := s.MockFilepathEvalSymlinks(map[string]string{})
	defer restore()
	restore = MockMountsPath("testdata/mounts-nvme")
	defer restore()
	restore = s.mockOsOpen(map[string]string{"/dev/nvme0n1": "testdata/disk.img"})
	defer restore()
	restore = MockSysfsPath(filepath.Join(s.UnpackTar(c, "testdata/sys.tar"), "sys"))
	defer restore()

	path, err := FilePathToDevicePath("/efi/ubuntu/shimx64.efi", FullPath)
	c.Check(err, IsNil)
	c.Check(path, DeepEquals, efi.DevicePath{
		&efi.ACPIDevicePathNode{HID: 0x0a0341d0},
		&efi.PCIDevicePathNode{Function: 0, Device: 0x1d},
		&efi.PCIDevicePathNode{Function: 0, Device: 0x0},
		&efi.NVMENamespaceDevicePathNode{NamespaceID: 1},
		&efi.HardDriveDevicePathNode{
			PartitionNumber: 1,
			PartitionStart:  34,
			PartitionSize:   64,
			Signature:       efi.GUIDHardDriveSignature(efi.MakeGUID(0xc7a2907e, 0xd8c9, 0x4a41, 0x8b99, [...]uint8{0x3e, 0xf3, 0x24, 0x5f, 0xaf, 0x2a})),
			MBRType:         efi.GPT},
		efi.FilePathDevicePathNode("\\EFI\\ubuntu\\shimx64.efi")})
}

func (s *filepathSuite) TestFilePathToDevicePathFullWithSymlink(c *C) {
	restore := s.MockFilepathEvalSymlinks(map[string]string{"/foo/bar/shimx64.efi": "/efi/ubuntu/shimx64.efi"})
	defer restore()
	restore = MockMountsPath("testdata/mounts-nvme")
	defer restore()
	restore = s.mockOsOpen(map[string]string{"/dev/nvme0n1": "testdata/disk.img"})
	defer restore()
	restore = MockSysfsPath(filepath.Join(s.UnpackTar(c, "testdata/sys.tar"), "sys"))
	defer restore()

	path, err := FilePathToDevicePath("/foo/bar/shimx64.efi", FullPath)
	c.Check(err, IsNil)
	c.Check(path, DeepEquals, efi.DevicePath{
		&efi.ACPIDevicePathNode{HID: 0x0a0341d0},
		&efi.PCIDevicePathNode{Function: 0, Device: 0x1d},
		&efi.PCIDevicePathNode{Function: 0, Device: 0x0},
		&efi.NVMENamespaceDevicePathNode{NamespaceID: 1},
		&efi.HardDriveDevicePathNode{
			PartitionNumber: 1,
			PartitionStart:  34,
			PartitionSize:   64,
			Signature:       efi.GUIDHardDriveSignature(efi.MakeGUID(0xc7a2907e, 0xd8c9, 0x4a41, 0x8b99, [...]uint8{0x3e, 0xf3, 0x24, 0x5f, 0xaf, 0x2a})),
			MBRType:         efi.GPT},
		efi.FilePathDevicePathNode("\\EFI\\ubuntu\\shimx64.efi")})
}
