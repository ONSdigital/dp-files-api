package files

import "github.com/ONSdigital/dp-kafka/v3/avro"

var AvroSchema = &avro.Schema{
	Definition: `{
			"type": "record",
			"name": "file-published",
			"fields": [
			  {"name": "path", "type": "string"},
			  {"name": "etag", "type": "string"},
			  {"name": "type", "type": "string"},
			  {"name": "sizeInBytes", "type": "string"}
			]
		  }`,
}

// FilePublished provides an avro structure for an image published event
type FilePublished struct {
	Path        string `avro:"path"`
	Type        string `avro:"type"`
	Etag        string `avro:"etag"`
	SizeInBytes string `avro:"sizeInBytes"`
}
