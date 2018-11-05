package fs

import (
	"os"
	"syscall"

	"github.com/icexin/gowasm/js"
)

type Constants struct {
	O_WRONLY int
	O_RDWR   int
	O_CREAT  int
	O_TRUNC  int
	O_APPEND int
	O_EXCL   int
}

func NewConstants() *Constants {
	return &Constants{
		O_WRONLY: os.O_WRONLY,
		O_RDWR:   os.O_RDWR,
		O_CREAT:  os.O_CREATE,
		O_TRUNC:  os.O_TRUNC,
		O_APPEND: os.O_APPEND,
		O_EXCL:   os.O_EXCL,
	}

}

type FS struct {
	Constants *Constants
}

func NewFS() *FS {
	return &FS{
		Constants: NewConstants(),
	}
}

func (f *FS) OpenSync(path string, flag, mode int64) (int, error) {
	return syscall.Open(path, int(flag), uint32(mode))
	return 0, js.ErrNoSys
}

type Stat struct {
	syscall.Stat_t
}

func (s *Stat) IsDirectory() bool {
	return s.Mode&syscall.S_IFMT == syscall.S_IFDIR
}

func (f *FS) FstatSync(fd int64) (*Stat, error) {
	var stat Stat
	err := syscall.Fstat(int(fd), &stat.Stat_t)
	return &stat, err
}

func (f *FS) WriteSync(fd int64, b []byte, offset, len int64) (int, error) {
	return syscall.Write(int(fd), b[offset:offset+len])
}

func (f *FS) ReadSync(fd int64, b []byte, offset, len int64) (int, error) {
	return syscall.Read(int(fd), b[offset:offset+len])
}

func (f *FS) CloseSync(fd int64) error {
	return syscall.Close(int(fd))
}

func init() {
	js.Register("Fs", NewFS())
}
