package log

type Metadata struct {
	TaskType       int    `yaml:"task_type"`
	LogDataPostfix string `yaml:"log_data_postfix"`
}

type LogMetadata struct {
	JobName      string     `yaml:"job_name"`
	LogDataLabel string     `yaml:"log_datalabel"`
	Metadatas    []Metadata `yaml:"metadatas"`
}

type LogMetaPubFormat struct {
	Timestamp   int64
	LogMetadata LogMetadata
}

func (logMeta *LogMetadata) AddMetadata(meta Metadata) {
	logMeta.Metadatas = append(logMeta.Metadatas, meta)
}
