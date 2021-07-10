package generate

import (
	_ "embed"
	"io"
	"os"
	"strings"
)

//go:embed tmpl/go_gen.tmpl
var gogenCode string

func GoGen(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, strings.NewReader(gogenCode)); err != nil {
		return err
	}

	return nil
}
