package packager

import (
	"fmt"
	"io"
	"os"
)

type Platform struct {
	OS, Arch string
}

func (p Platform) Key() string {
	return fmt.Sprintf("%s_%s", p.OS, p.Arch)
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
