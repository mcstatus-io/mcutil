package description

import (
	"fmt"
	"strings"
)

var (
	FormattingColorCodeLookupTable = map[rune]string{
		'0': "black",
		'1': "dark_blue",
		'2': "dark_green",
		'3': "dark_aqua",
		'4': "dark_red",
		'5': "dark_purple",
		'6': "gold",
		'7': "gray",
		'8': "dark_gray",
		'9': "blue",
		'a': "green",
		'b': "aqua",
		'c': "red",
		'd': "light_purple",
		'e': "yellow",
		'f': "white",
		'g': "minecoin_gold",
	}
	ColorNameLookupTable = map[string]rune{
		"black":         '0',
		"dark_blue":     '1',
		"dark_green":    '2',
		"dark_aqua":     '3',
		"dark_red":      '4',
		"dark_purple":   '5',
		"gold":          '6',
		"gray":          '7',
		"dark_gray":     '8',
		"blue":          '9',
		"green":         'a',
		"aqua":          'b',
		"red":           'c',
		"light_purple":  'd',
		"yellow":        'e',
		"white":         'f',
		"minecoin_gold": 'g',
	}
	HTMLColorLookupTable = map[string]string{
		"black":         "#000000",
		"dark_blue":     "#0000aa",
		"dark_green":    "#00aa00",
		"dark_aqua":     "#00aaaa",
		"dark_red":      "#aa0000",
		"dark_purple":   "#aa00aa",
		"gold":          "#ffaa00",
		"gray":          "#aaaaaa",
		"dark_gray":     "#555555",
		"blue":          "#5555ff",
		"green":         "#55ff55",
		"aqua":          "#55ffff",
		"red":           "#ff5555",
		"light_purple":  "#ff55ff",
		"yellow":        "#ffff55",
		"white":         "#ffffff",
		"minecoin_gold": "#ddd605",
	}
)

type MOTD struct {
	Tree  []FormatItem `json:"-"`
	Raw   string       `json:"raw"`
	Clean string       `json:"clean"`
	HTML  string       `json:"html"`
}

func ParseMOTD(desc interface{}) (res *MOTD, err error) {
	var tree []FormatItem

	switch v := desc.(type) {
	case string:
		{
			tree, err = parseString(v)

			break
		}
	case map[string]interface{}:
		{
			tree, err = parseString(parseChatObject(v))

			break
		}
	default:
		err = fmt.Errorf("unknown description type: %T", desc)
	}

	if err != nil {
		return
	}

	return &MOTD{
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

func parseChatObject(m map[string]interface{}) (result string) {
	if v, ok := m["color"].(string); ok {
		result += ParseColor(v).ToRaw()
	} else {
		result += White.ToRaw()
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
				result += parseChatObject(v2)
			}
		}
	}

	return
}

func parseString(s string) ([]FormatItem, error) {
	tree := make([]FormatItem, 0)

	item := FormatItem{
		Text:  "",
		Color: White,
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
				Color: White,
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
						Color: White,
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
