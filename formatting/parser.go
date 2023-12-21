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
func Parse(input interface{}) (*Result, error) {
	var (
		tree []Item = nil
		err  error
	)

	switch v := input.(type) {
	case string:
		{
			tree, err = parseString(v)

			break
		}
	case map[string]interface{}:
		{
			tree, err = parseString(parseChatObject(v, nil))

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

func parseChatObject(current map[string]interface{}, parent map[string]interface{}) (result string) {
	if v, ok := current["color"].(string); ok {
		result += colors.Parse(v).ToRaw()
	} else if v, ok := parent["color"].(string); ok {
		current["color"] = v

		result += colors.Parse(v).ToRaw()
	}

	if v, ok := current["bold"]; ok && parseBool(v) {
		result += decorators.Bold.ToRaw()
	} else if v, ok := parent["bold"]; ok && parseBool(v) {
		current["bold"] = v

		result += decorators.Bold.ToRaw()
	}

	if v, ok := current["italic"]; ok && parseBool(v) {
		result += decorators.Italic.ToRaw()
	} else if v, ok := parent["italic"]; ok && parseBool(v) {
		current["italic"] = v

		result += decorators.Italic.ToRaw()
	}

	if v, ok := current["underlined"]; ok && parseBool(v) {
		result += decorators.Underline.ToRaw()
	} else if v, ok := parent["underlined"]; ok && parseBool(v) {
		current["underlined"] = v

		result += decorators.Underline.ToRaw()
	}

	if v, ok := current["strikethrough"]; ok && parseBool(v) {
		result += decorators.Strikethrough.ToRaw()
	} else if v, ok := parent["strikethrough"]; ok && parseBool(v) {
		current["strikethrough"] = v

		result += decorators.Strikethrough.ToRaw()
	}

	if v, ok := current["obfuscated"]; ok && parseBool(v) {
		result += decorators.Obfuscated.ToRaw()
	} else if v, ok := parent["obfuscated"]; ok && parseBool(v) {
		current["obfuscated"] = v

		result += decorators.Obfuscated.ToRaw()
	}

	if text, ok := current["text"].(string); ok {
		result += text
	}

	if extra, ok := current["extra"].([]interface{}); ok {
		for _, v := range extra {
			if v2, ok := v.(map[string]interface{}); ok {
				result += parseChatObject(v2, current)
			}
		}
	}

	return
}

func parseString(s string) ([]Item, error) {
	var (
		tree []Item = make([]Item, 0)
		item Item   = Item{
			Text:       "",
			Decorators: make([]decorators.Decorator, 0),
		}
		r *strings.Reader = strings.NewReader(s)
	)

	for r.Len() > 0 {
		char, n, err := r.ReadRune()

		if err != nil {
			return nil, err
		}

		if n < 1 {
			break
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
					item.Color = &color
				} else {
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					item = Item{
						Text:       "",
						Color:      &color,
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
						Text:       "",
						Color:      item.Color,
						Decorators: append(item.Decorators, decorator),
					}
				}

				break
			}
		case 'r':
			{
				if len(item.Text) == 0 && len(item.Decorators) == 0 {
					break
				}

				tree = append(tree, item)

				item = Item{
					Text:       "",
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
