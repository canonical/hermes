package log

type Metadata struct {
	TaskType       int
	LogDataPostfix string
}

type LogMetadata struct {
	LogDataLabel string
	Metadatas    []Metadata
}

type LogMetaPubFormat struct {
	Timestamp   int64
	LogMetadata LogMetadata
}

func (logMeta *LogMetadata) AddMetadata(meta Metadata) {
	logMeta.Metadatas = append(logMeta.Metadatas, meta)
}
