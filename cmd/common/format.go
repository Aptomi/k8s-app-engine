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
	// Text is the plain text format (table) representation of object(s)
	Text = "text"
	// YAML format is just yaml marshaled object(s)
	YAML = "yaml"
	// JSON format is just json marshaled object(s)
	JSON = "json"
)

// Format returns string format for provided objects based on the output config
func Format(cfg *config.Client, list bool, objs ...runtime.Displayable) ([]byte, error) {
	switch strings.ToLower(cfg.Output) {
	case Text:
		return textMarshal(list, objs)
	case YAML:
		if list {
			return yaml.Marshal(objs)
		}
		return yaml.Marshal(objs[0])
	case JSON:
		if list {
			return json.Marshal(objs)
		}
		return json.Marshal(objs[0])
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
