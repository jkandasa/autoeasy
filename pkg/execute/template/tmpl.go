package template

import (
	"fmt"

	templateTY "github.com/jkandasa/autoeasy/pkg/types/template"
	fileUtils "github.com/jkandasa/autoeasy/pkg/utils/file"
)

func LoadTemplates(dir string) error {
	if !fileUtils.IsDirExists(dir) {
		return fmt.Errorf("template directory not found. dir:%s", dir)
	}

	files, err := fileUtils.ListFiles(dir)
	if err != nil {
		return err
	}

	// load templates
	for _, file := range files {
		if file.IsDir {
			continue
		}
		data, err := fileUtils.ReadFile(dir, file.Name)
		if err != nil {
			return err
		}
		tmpl := &templateTY.RawTemplate{
			Name:      file.Name,
			FileName:  file.FullPath,
			RawString: string(data),
		}

		err = add(tmpl)
		if err != nil {
			return err
		}
	}

	return nil
}
