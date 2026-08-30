package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/epheo/anytype-go"
)

// parseValue turns CLI text into the Go value NewPropertyLinkValue and
// FilterItem expect for a format. Empty text clears select/date and empties lists.
func parseValue(format anytype.PropertyFormat, raw string) (any, error) {
	switch format {
	case anytype.PropertyFormatNumber:
		return strconv.ParseFloat(raw, 64)
	case anytype.PropertyFormatCheckbox:
		return strconv.ParseBool(raw)
	case anytype.PropertyFormatSelect, anytype.PropertyFormatDate:
		if raw == "" {
			return nil, nil
		}
		return raw, nil
	case anytype.PropertyFormatMultiSelect, anytype.PropertyFormatFiles, anytype.PropertyFormatObjects:
		if raw == "" {
			return []string{}, nil
		}
		return strings.Split(raw, ","), nil
	}
	return raw, nil
}

// propertyFormats maps every property key in the space to its format, so
// users type key=value without stating the type.
func propertyFormats(s *session, sc anytype.SpaceContext) (map[string]anytype.PropertyFormat, error) {
	formats := map[string]anytype.PropertyFormat{}
	for p, err := range sc.Properties().All(s.ctx) {
		if err != nil {
			return nil, fmt.Errorf("list properties: %w", err)
		}
		formats[p.Key] = p.Format
	}
	return formats, nil
}

func parsePropertyValues(kvs []string, formats map[string]anytype.PropertyFormat) ([]anytype.PropertyLinkValue, error) {
	out := make([]anytype.PropertyLinkValue, 0, len(kvs))
	for _, kv := range kvs {
		key, raw, err := splitKV(kv)
		if err != nil {
			return nil, err
		}
		format, ok := formats[key]
		if !ok {
			return nil, fmt.Errorf("unknown property key %q", key)
		}
		v, err := parseValue(format, raw)
		if err != nil {
			return nil, fmt.Errorf("property %s: %w", key, err)
		}
		link, err := anytype.NewPropertyLinkValue(key, format, v)
		if err != nil {
			return nil, err
		}
		out = append(out, link)
	}
	return out, nil
}

func propertyValues(s *session, sc anytype.SpaceContext, kvs []string) ([]anytype.PropertyLinkValue, error) {
	if len(kvs) == 0 {
		return nil, nil
	}
	formats, err := propertyFormats(s, sc)
	if err != nil {
		return nil, err
	}
	return parsePropertyValues(kvs, formats)
}

var filterConditions = []anytype.FilterCondition{
	anytype.FilterConditionEq, anytype.FilterConditionNe,
	anytype.FilterConditionIn, anytype.FilterConditionNin,
	anytype.FilterConditionContains, anytype.FilterConditionNContains,
	anytype.FilterConditionGt, anytype.FilterConditionLt,
	anytype.FilterConditionGte, anytype.FilterConditionLte,
	anytype.FilterConditionAll,
	anytype.FilterConditionEmpty, anytype.FilterConditionNEmpty,
}

// parseFilters reads "key:condition[:value]"; empty and nempty take no value.
// System keys such as "name" are absent from the property list, so an unknown
// key is treated as text and left for the API to validate.
func parseFilters(specs []string, formats map[string]anytype.PropertyFormat) ([]anytype.FilterItem, error) {
	items := make([]anytype.FilterItem, 0, len(specs))
	for _, spec := range specs {
		parts := strings.SplitN(spec, ":", 3)
		if len(parts) < 2 || parts[0] == "" {
			return nil, fmt.Errorf("expected key:condition[:value], got %q", spec)
		}
		key, cond := parts[0], anytype.FilterCondition(parts[1])
		if !validCondition(cond) {
			return nil, fmt.Errorf("filter %s: unknown condition %q (%s)", key, cond, joinConsts(filterConditions...))
		}
		format, ok := formats[key]
		if !ok {
			format = anytype.PropertyFormatText
		}
		if cond == anytype.FilterConditionEmpty || cond == anytype.FilterConditionNEmpty {
			items = append(items, anytype.EmptyFilter(key, cond))
			continue
		}
		if len(parts) != 3 {
			return nil, fmt.Errorf("filter %s: condition %s needs a value", key, cond)
		}
		v, err := parseValue(format, parts[2])
		if err != nil {
			return nil, fmt.Errorf("filter %s: %w", key, err)
		}
		items = append(items, anytype.FilterItem{Key: key, Format: format, Condition: cond, Value: v})
	}
	return items, nil
}

func validCondition(c anytype.FilterCondition) bool {
	for _, known := range filterConditions {
		if c == known {
			return true
		}
	}
	return false
}
