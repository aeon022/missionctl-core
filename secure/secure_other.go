//go:build !unix

package secure

func Init(tool string)   {}
func Tighten(dir string) {}
