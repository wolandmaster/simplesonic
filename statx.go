package main

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

type Statx_t struct {
	Mask            uint32
	Blksize         uint32
	Attributes      uint64
	Nlink           uint32
	Uid             uint32
	Gid             uint32
	Mode            uint16
	_               [1]uint16
	Ino             uint64
	Size            uint64
	Blocks          uint64
	Attributes_mask uint64
	Atime           StatxTimestamp
	Btime           StatxTimestamp
	Ctime           StatxTimestamp
	Mtime           StatxTimestamp
	Rdev_major      uint32
	Rdev_minor      uint32
	Dev_major       uint32
	Dev_minor       uint32
	Mnt_id          uint64
	_               uint64
	_               [12]uint64
}

type StatxTimestamp struct {
	Sec  int64
	Nsec uint32
	_    int32
}

var (
	atFdCwd           = -0x64
	atSymlinkNofollow = 0x100
	statxAll          = 0xfff
	sysStatxArch      = map[string]int{
		"386": 383, "amd64": 332, "arm": 397, "arm64": 291, "ppc": 383, "ppc64": 383, "ppc64le": 383, "loong64": 291,
		"s390x": 379, "sparc64": 360, "riscv64": 291, "mips": 4366, "mipsle": 4366, "mips64": 5326, "mips64le": 5326,
	}
)

func Statx(filename string) (*Statx_t, error) {
	if sysStatx, ok := sysStatxArch[runtime.GOARCH]; ok && runtime.GOOS == "linux" {
		var statx Statx_t
		filenamePtr := &append([]byte(filename), 0x0)[0]
		if _, _, errno := syscall.Syscall6(uintptr(sysStatx), uintptr(atFdCwd), uintptr(unsafe.Pointer(filenamePtr)),
			uintptr(atSymlinkNofollow), uintptr(statxAll), uintptr(unsafe.Pointer(&statx)), 0); errno != 0 {
			return nil, fmt.Errorf("statx errno %d", errno)
		}
		return &statx, nil
	}
	return nil, nil
}
