package common

import (
	"encoding/json"
	"fmt"
	"github.com/Aptomi/aptomi/pkg/config"
	"github.com/Aptomi/aptomi/pkg/runtime"
	"github.com/gosuri/uitable"
	"gopkg.in/yaml.v2"
	"strings"
)

const (
	Text = "text"
	YAML = "yaml"
	JSON = "json"
)

func Format(cfg *config.Client, list bool, objs ...runtime.Displayable) ([]byte, error) {
	switch strings.ToLower(cfg.Output) {
	case Text:
		return textMarshal(list, objs)
	case YAML:
		if list {
			return yaml.Marshal(objs)
		} else {
			return yaml.Marshal(objs[0])
		}
	case JSON:
		if list {
			return json.Marshal(objs)
		} else {
			return json.Marshal(objs[0])
		}
	}

	panic(fmt.Sprintf("%s output format not supported", cfg.Output))
}

func textMarshal(list bool, objs []runtime.Displayable) ([]byte, error) {
	table := uitable.New()
	table.MaxColWidth = 120
	table.Wrap = true

	defaultColumns := objs[0].GetDefaultColumns()
	columns := make([]interface{}, len(defaultColumns))

	for idx := range defaultColumns {
		columns[idx] = defaultColumns[idx]
	}

	// fill table headers
	table.AddRow(columns...)

	for _, obj := range objs {
		allColumns := obj.AsColumns()

		for idx, column := range defaultColumns {
			columns[idx] = allColumns[column]
		}

		table.AddRow(columns...)
	}

	return table.Bytes(), nil
}
