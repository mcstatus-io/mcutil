package description

import (
	"fmt"
	"strings"
)

// Formatting is a parsed structure of any Minecraft string encoded with color codes or format codes
type Formatting struct {
	Tree  []FormatItem `json:"-"`
	Raw   string       `json:"raw"`
	Clean string       `json:"clean"`
	HTML  string       `json:"html"`
}

// ParseFormatting parses the formatting of any string or Chat object
func ParseFormatting(desc interface{}, defaultColor ...Color) (res *Formatting, err error) {
	var tree []FormatItem

	defColor := White

	if len(defaultColor) > 0 {
		defColor = defaultColor[0]
	}

	switch v := desc.(type) {
	case string:
		{
			tree, err = parseString(v, defColor)

			break
		}
	case map[string]interface{}:
		{
			tree, err = parseString(parseChatObject(v, defColor), defColor)

			break
		}
	default:
		err = fmt.Errorf("description: unknown description type: %T", desc)
	}

	if err != nil {
		return
	}

	return &Formatting{
		Tree:  tree,
		Raw:   toRaw(tree),
		Clean: toClean(tree),
		HTML:  toHTML(tree),
	}, nil
}

func toRaw(tree []FormatItem) (result string) {
	for _, v := range tree {
		result += v.Raw()
	}

	return
}

func toClean(tree []FormatItem) (result string) {
	for _, v := range tree {
		result += v.Clean()
	}

	return
}

func toHTML(tree []FormatItem) (result string) {
	result = "<span>"

	for _, v := range tree {
		result += v.HTML()
	}

	result += "</span>"

	return
}

func parseChatObject(m map[string]interface{}, defaultColor Color) (result string) {
	if v, ok := m["color"].(string); ok {
		result += ParseColor(v).ToRaw()
	} else {
		result += defaultColor.ToRaw()
	}

	if v, ok := m["bold"]; ok && parseBool(v) {
		result += "\u00A7l"
	}

	if v, ok := m["italic"]; ok && parseBool(v) {
		result += "\u00A7o"
	}

	if v, ok := m["underlined"]; ok && parseBool(v) {
		result += "\u00A7n"
	}

	if v, ok := m["strikethrough"]; ok && parseBool(v) {
		result += "\u00A7m"
	}

	if v, ok := m["obfuscated"]; ok && parseBool(v) {
		result += "\u00A7k"
	}

	if text, ok := m["text"].(string); ok {
		result += text
	}

	if extra, ok := m["extra"].([]interface{}); ok {
		for _, v := range extra {
			if v2, ok := v.(map[string]interface{}); ok {
				result += parseChatObject(v2, defaultColor)
			}
		}
	}

	return
}

func parseString(s string, defaultColor Color) ([]FormatItem, error) {
	tree := make([]FormatItem, 0)

	item := FormatItem{
		Text:  "",
		Color: defaultColor,
	}

	r := strings.NewReader(s)

	for r.Len() > 0 {
		char, n, err := r.ReadRune()

		if err != nil {
			return nil, err
		}

		if n < 1 {
			break
		}

		if char == '\n' {
			tree = append(tree, item)

			item = FormatItem{
				Text:  "\n",
				Color: defaultColor,
			}

			continue
		}

		if char != '\u00A7' {
			item.Text += string(char)

			continue
		}

		code, n, err := r.ReadRune()

		if err != nil {
			return nil, err
		}

		if n < 1 {
			break
		}

		// Color code
		{
			if code >= '0' && code <= 'g' {
				newColor := ParseColor(string(code))

				if item.Obfuscated || item.Bold || item.Strikethrough || item.Underline || item.Italic || newColor != item.Color {
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item = FormatItem{
						Text:  "",
						Color: newColor,
					}
				} else {
					item.Color = newColor
				}

				continue
			}
		}

		// Formatting code
		{
			switch code {
			case 'k':
				{
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item.Text = ""
					item.Obfuscated = true
				}
			case 'l':
				{
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item.Text = ""
					item.Bold = true
				}
			case 'm':
				{
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item.Text = ""
					item.Strikethrough = true
				}
			case 'n':
				{
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item.Text = ""
					item.Underline = true
				}
			case 'o':
				{
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item.Text = ""
					item.Italic = true
				}
			case 'r':
				{
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item = FormatItem{
						Text:  "",
						Color: defaultColor,
					}
				}
			}
		}
	}

	tree = append(tree, item)

	return tree, nil
}

func parseBool(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.ToLower(v) == "true"
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
		return v == 1
	default:
		return false
	}
}
