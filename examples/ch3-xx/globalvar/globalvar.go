// Copyright © 2017 ChaiShushan <chaishushan{AT}gmail.com>.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// 汇编中访问Go中定义的全局变量

package globalvar

var gopkgValue int = 42

type PkgInfo struct {
	V0 byte
	V1 uint16
	V2 int32
	V3 int64
	V4 bool
	V5 bool
}

var gInfo PkgInfo

func init() {
	gInfo.V0 = 101
	gInfo.V1 = 102
	gInfo.V2 = 103
	gInfo.V3 = 104
	gInfo.V4 = true
	gInfo.V5 = false
}

func GetPkgValue() int

func GetPkgInfo() PkgInfo
