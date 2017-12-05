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
		return yaml.Marshal(objs)
	case JSON:
		return json.Marshal(objs)
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
