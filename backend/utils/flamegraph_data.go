package utils

import (
	"encoding/json"
	"os"
)

type FlameGraphData struct {
	Name     string
	Value    int64
	Children map[string]*FlameGraphData
}

func NewFlameGraphData() *FlameGraphData {
	return &FlameGraphData{
		Name:     "root",
		Value:    0,
		Children: make(map[string]*FlameGraphData),
	}
}

func (data *FlameGraphData) Add(stack *[]string, idx int, val int64) {
	data.Value += val
	if idx < 0 {
		return
	}
	name := (*stack)[idx]
	ptr, isExist := data.Children[name]
	if !isExist {
		ptr = &FlameGraphData{
			Name:     name,
			Value:    0,
			Children: make(map[string]*FlameGraphData),
		}
		data.Children[name] = ptr
	}
	ptr.Add(stack, idx-1, val)
}

func (data *FlameGraphData) MarshalJSON() ([]byte, error) {
	children := make([]FlameGraphData, len(data.Children))
	for _, child := range data.Children {
		children = append(children, *child)
	}

	return json.Marshal(&struct {
		Name     string           `json:"name"`
		Value    int64            `json:"value"`
		Children []FlameGraphData `json:"children"`
	}{
		Name:     data.Name,
		Value:    data.Value,
		Children: children,
	})
}

func (data *FlameGraphData) WriteToFile(path string) error {
	bytes, err := data.MarshalJSON()
	if err != nil {
		return err
	}
	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	if _, err = fp.WriteString(string(bytes)); err != nil {
		return err
	}
	return nil
}
