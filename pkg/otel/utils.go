package otel

import (
	"fmt"

	"go.opentelemetry.io/otel/attribute"
)

func ParseAttributes(attrsMap map[string]interface{}) []attribute.KeyValue {
	var attrs []attribute.KeyValue
	for k, v := range attrsMap {
		switch v.(type) {
		case bool:
			attrs = append(attrs, attribute.Bool(k, v.(bool)))
		case string:
			attrs = append(attrs, attribute.String(k, v.(string)))
		case int64:
			attrs = append(attrs, attribute.Int64(k, v.(int64)))
		case float64:
			attrs = append(attrs, attribute.Float64(k, v.(float64)))
		case []float64:
			attrs = append(attrs, attribute.Float64Slice(k, v.([]float64)))
		case []int64:
			attrs = append(attrs, attribute.Int64Slice(k, v.([]int64)))
		case []string:
			attrs = append(attrs, attribute.StringSlice(k, v.([]string)))
		case []bool:
			attrs = append(attrs, attribute.BoolSlice(k, v.([]bool)))
		default:
			attrs = append(attrs, attribute.String(k, fmt.Sprintf("%v", v)))
		}
	}
	return attrs
}
