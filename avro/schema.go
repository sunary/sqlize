package avro

// Field ...
type Field struct {
	Name    string      `json:"name"`
	Type    interface{} `json:"type"`
	Default interface{} `json:"default,omitempty"`
}

// RecordSchema ...
type RecordSchema struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	//shouldn't set directly to this field, only read and for json export
	Fields      []Field `json:"fields"`
	ConnectName string  `json:"connect.name"`
}

func newRecordSchema(namespace, connectName string) *RecordSchema {
	return &RecordSchema{
		Type:        "record",
		Name:        "envelop",
		Namespace:   namespace,
		ConnectName: connectName,
		Fields: []Field{
			{
				Name:    "before",
				Default: nil,
			},
			{
				Name: "after",
				Type: []string{
					"null",
					"Value",
				},
				Default: nil,
			},
			{
				Name: "op",
				Type: "string",
			},
			{
				Name: "ts_ms",
				Type: []interface{}{
					"null",
					"long",
				},
				Default: nil,
			},
			{
				Name: "transaction",
				Type: []interface{}{
					"null",
					RecordSchema{
						Type:      "record",
						Name:      "ConnectDefault",
						Namespace: "io.confluent.connect.avro",
						Fields: []Field{
							{
								Name: "id",
								Type: "string",
							},
							{
								Name: "total_order",
								Type: "long",
							},
							{
								Name: "data_collection_order",
								Type: "long",
							},
						},
					},
				},
				Default: nil,
			},
		},
	}
}
