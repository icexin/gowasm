package js

var (
	ErrNotfound        = NewException("EEXIS", "not found")
	ErrNoSys           = NewException("ENOSYS", "not implemention")
	ErrInvalidArgument = NewException("EINVAL", "invalid argument")
)

type Exception struct {
	Code    string
	Message string
}

func NewException(code, msg string) error {
	return &Exception{
		Code:    code,
		Message: msg,
	}
}

func (e *Exception) Error() string {
	return e.Message
}

func Array() []interface{} {
	return []interface{}{}
}

func Uint8Array(b []byte, offset int64, len int64) []byte {
	return b[offset : offset+len]
}

type Memory struct {
	memfunc func() []byte
}

func NewMemory(f func() []byte) *Memory {
	return &Memory{
		memfunc: f,
	}
}

func (m *Memory) Get(name string) (interface{}, bool) {
	return m.memfunc(), true
}

func init() {
	Register("Array", Array)
	Register("Uint8Array", Uint8Array)
}
