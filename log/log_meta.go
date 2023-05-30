package log

type Metadata struct {
	TaskType       int    `yaml:"task_type"`
	LogDataPostfix string `yaml:"log_data_postfix"`
}

type LogMetadata struct {
	LogDataLabel string     `yaml:"data_label"`
	Metadatas    []Metadata `yaml:"metadatas"`
}

func (logMeta *LogMetadata) AddMetadata(meta Metadata) {
	logMeta.Metadatas = append(logMeta.Metadatas, meta)
}
