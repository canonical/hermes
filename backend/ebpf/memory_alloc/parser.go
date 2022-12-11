package ebpf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type MemoryParser struct {
	Recs map[uint64]DataRecord
}

func GetParser() (*MemoryParser, error) {
	return &MemoryParser{
		Recs: make(map[uint64]DataRecord),
	}, nil
}

func (parser *MemoryParser) getStackCollapsedData(pid uint32, comm string, allocRec *AllocRecord) string {
	data := fmt.Sprintf("%s;", comm)
	for i := len(allocRec.CallchainInsts) - 1; i >= 0; i-- {
		data += fmt.Sprintf("%s;", allocRec.CallchainInsts[i])
	}
	data = data[:len(data)-1]
	data += fmt.Sprintf(" %d", pid)
	return data
}

func (parser *MemoryParser) writeStackCollapsedData(outputPath string) error {
	fp, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	idx := 0
	for tgidPid, rec := range parser.Recs {
		idx++
		var data string
		for _idx, allocRec := range rec.AllocRecs {
			data += parser.getStackCollapsedData(uint32(tgidPid), rec.Comm, &allocRec)
			if _idx != len(rec.AllocRecs)-1 {
				data += "\n"
			}
		}
		if idx == len(parser.Recs) {
			data += "\n"
		}
		if _, err = fp.WriteString(data); err != nil {
			return err
		}

	}
	return nil
}

func (parser *MemoryParser) Parse(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &parser.Recs)
	if err != nil {
		return err
	}

	return parser.writeStackCollapsedData(path + string(".stack.collapsed"))
}
