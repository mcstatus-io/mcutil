package formatting

import (
	"fmt"
	"strings"

	"github.com/mcstatus-io/mcutil/v2/formatting/colors"
	"github.com/mcstatus-io/mcutil/v2/formatting/decorators"
)

// Result is a parsed structure of any Minecraft string encoded with color codes or format codes
type Result struct {
	Tree  []Item `json:"-"`
	Raw   string `json:"raw"`
	Clean string `json:"clean"`
	HTML  string `json:"html"`
}

// Parse parses the formatting of any string or Chat object
func Parse(input interface{}, defaultColor ...colors.Color) (*Result, error) {
	var (
		tree []Item = nil
		err  error
	)

	defColor := colors.White

	if len(defaultColor) > 0 {
		defColor = defaultColor[0]
	}

	switch v := input.(type) {
	case string:
		{
			tree, err = parseString(v, defColor)

			break
		}
	case map[string]interface{}:
		{
			tree, err = parseString(parseChatObject(v, nil, defColor), defColor)

			break
		}
	default:
		err = fmt.Errorf("unknown formatting type: %T", input)
	}

	if err != nil {
		return nil, err
	}

	return &Result{
		Tree:  tree,
		Raw:   toRaw(tree),
		Clean: toClean(tree),
		HTML:  toHTML(tree),
	}, nil
}

func parseChatObject(m map[string]interface{}, p map[string]interface{}, defaultColor colors.Color) (result string) {
	if v, ok := m["color"].(string); ok {
		result += colors.Parse(v).ToRaw()
	} else if v, ok := p["color"].(string); ok {
		m["color"] = v
		result += colors.Parse(v).ToRaw()
	} else {
		result += defaultColor.ToRaw()
	}

	if v, ok := m["bold"]; ok && parseBool(v) {
		result += decorators.Bold.ToRaw()
	} else if v, ok := p["bold"]; ok && parseBool(v) {
		m["bold"] = v
		result += decorators.Bold.ToRaw()
	}

	if v, ok := m["italic"]; ok && parseBool(v) {
		result += decorators.Italic.ToRaw()
	} else if v, ok := p["italic"]; ok && parseBool(v) {
		m["italic"] = v
		result += decorators.Italic.ToRaw()
	}

	if v, ok := m["underlined"]; ok && parseBool(v) {
		result += decorators.Underline.ToRaw()
	} else if v, ok := p["underlined"]; ok && parseBool(v) {
		m["underlined"] = v
		result += decorators.Underline.ToRaw()
	}

	if v, ok := m["strikethrough"]; ok && parseBool(v) {
		result += decorators.Strikethrough.ToRaw()
	} else if v, ok := p["strikethrough"]; ok && parseBool(v) {
		m["strikethrough"] = v
		result += decorators.Strikethrough.ToRaw()
	}

	if v, ok := m["obfuscated"]; ok && parseBool(v) {
		result += decorators.Obfuscated.ToRaw()
	} else if v, ok := p["obfuscated"]; ok && parseBool(v) {
		m["obfuscated"] = v
		result += decorators.Obfuscated.ToRaw()
	}

	if text, ok := m["text"].(string); ok {
		result += text
	} else if translate, ok := m["translate"].(string); ok {
		result += translate
	}

	if extra, ok := m["extra"].([]interface{}); ok {
		for _, v := range extra {
			if v2, ok := v.(map[string]interface{}); ok {
				result += parseChatObject(v2, m, defaultColor)
			}
		}
	}

	return
}

func parseString(s string, defaultColor colors.Color) ([]Item, error) {
	tree := make([]Item, 0)

	item := Item{
		Text:       "",
		Color:      defaultColor,
		Decorators: make([]decorators.Decorator, 0),
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

			item = Item{
				Text:       "\n",
				Color:      defaultColor,
				Decorators: make([]decorators.Decorator, 0),
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

		switch code {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f', 'g':
			{
				color := colors.Parse(code)

				if len(item.Text) == 0 && len(item.Decorators) == 0 {
					item.Color = color
				} else {
					tree = append(tree, item)

					item = Item{
						Text:       "",
						Color:      color,
						Decorators: make([]decorators.Decorator, 0),
					}
				}

				break
			}
		case 'k', 'l', 'm', 'n', 'o':
			{
				decorator := decorators.Parse(code)

				if len(item.Text) == 0 {
					item.Decorators = append(item.Decorators, decorator)
				} else {
					tree = append(tree, item)

					item = Item{
						Text:  "",
						Color: defaultColor,
						Decorators: []decorators.Decorator{
							decorator,
						},
					}
				}

				break
			}
		case 'r':
			{
				if len(item.Text) == 0 && item.Color == defaultColor && len(item.Decorators) == 0 {
					break
				}

				tree = append(tree, item)

				item = Item{
					Text:       "",
					Color:      defaultColor,
					Decorators: make([]decorators.Decorator, 0),
				}

				break
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
