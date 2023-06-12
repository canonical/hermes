package main

import (
	pb "hermes/proto"
	"io/ioutil"
)

type ContentParser struct {
	dir string
}

func NewContentParser(viewDir string) *ContentParser {
	return &ContentParser{
		dir: viewDir,
	}
}

func (parser *ContentParser) GetRoutines() *pb.Routines {
	var routines []string

	files, err := ioutil.ReadDir(parser.dir)
	if err != nil {
		return nil
	}
	for _, file := range files {
		routines = append(routines, file.Name())
	}
	return &pb.Routines{
		Routines: routines,
	}
}
