package asmpkg

func CallCAdd_SystemV_ABI(cfun uintptr, a, b int64) int64
func CallCAdd_Win64_ABI(cfun uintptr, a, b int64) int64

func SyscallWrite_Darwin(fd int, msg string) int
func SyscallWrite_Linux(fd int, msg string) int
func SyscallWrite_Windows(fd int, msg string) int

func CopySlice_AVX2(dst, src []byte, len int)
