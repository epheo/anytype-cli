package output

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/epheo/anytype-go"
)

// PropertyValue renders whichever typed field the property format uses.
// Zero-valued numbers and unchecked boxes print as such rather than "empty".
func PropertyValue(p anytype.Property) string {
	switch p.Format {
	case anytype.PropertyFormatText:
		return p.Text
	case anytype.PropertyFormatNumber:
		return strconv.FormatFloat(p.Number, 'f', -1, 64)
	case anytype.PropertyFormatCheckbox:
		return strconv.FormatBool(p.Checkbox)
	case anytype.PropertyFormatDate:
		return p.Date
	case anytype.PropertyFormatURL:
		return p.URL
	case anytype.PropertyFormatEmail:
		return p.Email
	case anytype.PropertyFormatPhone:
		return p.Phone
	case anytype.PropertyFormatFiles:
		return strings.Join(p.Files, ", ")
	case anytype.PropertyFormatObjects:
		return strings.Join(p.Objects, ", ")
	case anytype.PropertyFormatSelect:
		if p.Select != nil {
			return p.Select.Name
		}
		return ""
	case anytype.PropertyFormatMultiSelect:
		names := make([]string, len(p.MultiSelect))
		for i, t := range p.MultiSelect {
			names[i] = t.Name
		}
		return strings.Join(names, ", ")
	}
	return fmt.Sprintf("<%s>", p.Format)
}

func Icon(icon *anytype.Icon) string {
	if icon == nil {
		return ""
	}
	switch icon.Format {
	case anytype.IconFormatEmoji:
		return icon.Emoji
	case anytype.IconFormatFile:
		return "file:" + icon.File
	case anytype.IconFormatIcon:
		if icon.Color != "" {
			return icon.Name + " (" + string(icon.Color) + ")"
		}
		return icon.Name
	}
	return ""
}
