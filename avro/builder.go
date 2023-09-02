package avro

import (
	"strconv"
	"strings"

	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/types"
	"github.com/sunary/sqlize/element"
)

// NewArvoSchema ...
func NewArvoSchema(table element.Table) *RecordSchema {
	fields := buildFieldsFromTable(table)
	record := newRecordSchema(table.Name, table.Name)
	record.Name = table.Name
	//set payload , on before field
	record.Fields[0].Type = []interface{}{
		"null",
		RecordSchema{
			Name:   "Value",
			Type:   "record",
			Fields: fields,
		},
	}

	return record
}

func buildFieldsFromTable(table element.Table) []Field {
	fields := []Field{}

	for _, col := range table.Columns {
		field := Field{
			Name: col.Name,
		}
		if col.HasDefaultValue() {
			field.Type = []interface{}{"null", getAvroType(col)}
			field.Default = nil
		} else {
			field.Type = getAvroType(col)
		}
		fields = append(fields, field)
	}

	return fields
}

func getAvroType(col element.Column) interface{} {
	switch col.GetType() {
	//case mysql.TypeBit, mysql.TypeBlob, mysql.TypeTinyBlob, mysql.TypeMediumBlob, mysql.TypeLongBlob:
	//	return "[]byte"

	case mysql.TypeTiny:
		return "bool"

	case mysql.TypeEnum:
		return map[string]interface{}{
			"type":            "string",
			"connect.version": 1,
			"connect.parameters": map[string]string{
				"allowed": strings.Join(col.MysqlType.Elems, ","),
			},
			"connect.default": "init",
			"connect.name":    "io.debezium.data.Enum",
		}
	}

	switch col.MysqlType.EvalType() {
	case types.ETInt:
		return "int"

	case types.ETDecimal:
		displayFlen, displayDecimal := col.MysqlType.Flen, col.MysqlType.Decimal
		return map[string]interface{}{
			"type":            "bytes",
			"scale":           displayDecimal,
			"precision":       displayFlen,
			"connect.version": 1,
			"connect.parameters": map[string]string{
				"scale":                     strconv.Itoa(displayDecimal),
				"connect.decimal.precision": strconv.Itoa(displayFlen),
			},
			"connect.name": "org.apache.kafka.connect.data.Decimal",
			"logicalType":  "decimal",
		}

	case types.ETReal:
		return "float64"

	case types.ETDatetime, types.ETTimestamp:
		return map[string]interface{}{
			"type":            "string",
			"connect.version": 1,
			"connect.default": "1970-01-01T00:00:00Z",
			"connect.name":    "io.debezium.time.ZonedTimestamp",
		}

	case types.ETJson:
		return map[string]interface{}{
			"type":            "string",
			"connect.version": 1,
			"connect.name":    "io.debezium.data.Json",
		}

	case types.ETString:
		return "string"

	default:
		return "string"
	}
}
