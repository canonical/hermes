package plugin

import (
	"encoding/json"
	"net/http"
)

type FlamebearerData struct {
	Names    []string  `json:"names"`
	Levels   [][]int64 `json:"levels"`
	NumTicks int64     `json:"numTicks"`
	MaxSelf  int64     `json:"maxSelf"`
}

type Metadata struct {
	Format     string `json:"format"`
	SampleRate int    `json:"sampleRate"`
	Units      string `json:"units"`
	Name       string `json:"name"`
	AppName    string `json:"appName"`
}

type ProfileData struct {
	Version         int             `json:"version"`
	FlamebearerData FlamebearerData `json:"flamebearer"`
	Metadata        Metadata        `json:"metadata"`
}

type FlamegraphData struct {
	Name     string           `json:"name"`
	Value    int64            `json:"value"`
	Children []FlamegraphData `json:"children"`
}

func _convertToFlamebearer(flamegraphs [][]*FlamegraphData, flamebearerData *FlamebearerData) {
	_flamegraphs := [][]*FlamegraphData{}
	hasChild := false
	levels := []int64{}

	parentLevel := []int64{}
	if len(flamebearerData.Levels) != 0 {
		parentLevel = flamebearerData.Levels[len(flamebearerData.Levels)-1]
	}

	var padding int64 = 0
	for i, flamegraph := range flamegraphs {
		if len(parentLevel) != 0 {
			padding += parentLevel[i*4] + parentLevel[i*4+2]
		}
		for j, _flamegraph := range flamegraph {
			if _flamegraph.Value == 0 {
				continue
			}

			__flamegraphs := []*FlamegraphData{}
			selfVal := _flamegraph.Value
			for k, child := range _flamegraph.Children {
				hasChild = true
				__flamegraphs = append(__flamegraphs, &flamegraphs[i][j].Children[k])
				selfVal -= child.Value
			}

			flamebearerData.Names = append(flamebearerData.Names, _flamegraph.Name)
			if selfVal > flamebearerData.MaxSelf {
				flamebearerData.MaxSelf = selfVal
			}

			level := []int64{0, _flamegraph.Value, selfVal, int64(len(flamebearerData.Names) - 1)}
			if padding != 0 {
				level[0] = padding
				padding = 0
			}
			levels = append(levels, level...)

			_flamegraphs = append(_flamegraphs, __flamegraphs)
		}
	}
	flamebearerData.Levels = append(flamebearerData.Levels, levels)
	if hasChild {
		_convertToFlamebearer(_flamegraphs, flamebearerData)
	}
}

func convertToFlamebearer(flamegraphData *FlamegraphData, flamebearerData *FlamebearerData) {
	flamegraphs := [][]*FlamegraphData{
		{flamegraphData},
	}

	flamebearerData.NumTicks = int64(flamegraphData.Value)
	_convertToFlamebearer(flamegraphs, flamebearerData)
}

func HandleResResp(resp []byte, status *int, respBody *[]byte) error {
	var flamegraphData FlamegraphData
	if err := json.Unmarshal(resp, &flamegraphData); err != nil {
		return err
	}

	profileData := ProfileData{
		Version: 1,
		Metadata: Metadata{
			Format:     "single",
			SampleRate: 10,
			Units:      "samples",
			Name:       "cpu profile",
			AppName:    "hermes.cpu_profile",
		},
	}
	convertToFlamebearer(&flamegraphData, &profileData.FlamebearerData)

	_respBody, err := json.Marshal(&profileData)
	if err != nil {
		return err
	}

	*respBody = _respBody
	*status = http.StatusOK
	return nil
}
