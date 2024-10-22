package formatting

import (
	"fmt"
	"strings"

	"github.com/mcstatus-io/mcutil/v4/formatting/colors"
	"github.com/mcstatus-io/mcutil/v4/formatting/decorators"
)

// Result is a parsed structure of any Minecraft string encoded with color codes or format codes.
type Result struct {
	Tree  []Item `json:"-"`
	Raw   string `json:"raw"`
	Clean string `json:"clean"`
	HTML  string `json:"html"`
}

// Parse parses the formatting of any string or Chat object.
func Parse(input interface{}) (*Result, error) {
	tree, err := parseAny(input, nil)

	if err != nil {
		return nil, err
	}

	tree = cleanupTree(tree)

	return &Result{
		Tree:  tree,
		Raw:   toRaw(tree),
		Clean: toClean(tree),
		HTML:  toHTML(tree),
	}, nil
}

func parseAny(input interface{}, parent map[string]interface{}) ([]Item, error) {
	result := make([]Item, 0)

	switch value := input.(type) {
	case map[string]interface{}:
		{
			// Merge the properties from the parent component into this one.
			current := mergeChatProperties(parent, value)

			// Only add a new item if there is text in this component, since it is useless
			// to push a new item that has no displayable text.
			if text, ok := current["text"].(string); ok {
				children, err := parseString(text, current)

				if err != nil {
					return nil, err
				}

				result = append(result, children...)
			}

			// Iterate over the child components and pass the current component to
			// them so they can inherit our properties.
			if extra, ok := current["extra"].([]interface{}); ok {
				for _, child := range extra {
					extraChildren, err := parseAny(child, current)

					if err != nil {
						return nil, err
					}

					result = append(result, extraChildren...)
				}
			}

			break
		}
	case string:
		{
			children, err := parseString(value, parent)

			if err != nil {
				return nil, err
			}

			// It is possible that a child value of an "extra" property on a chat component may be a string and
			// not another chat component, so the parents' properties must be inherited on this string as well.
			// See: https://github.com/mcstatus-io/mcutil/issues/6
			result = append(result, children...)

			break
		}
	default:
		return nil, fmt.Errorf("format: unexpected input type: %+v", input)
	}

	// Remove component items with an empty text, as they do not do anything to the result other
	// than just wasting space.
	for i := 0; i < len(result); i++ {
		if len(result[i].Text) > 0 {
			continue
		}

		result = append(result[0:i], result[i+1:]...)

		i--
	}

	return result, nil
}

func parseString(text string, props map[string]interface{}) ([]Item, error) {
	if strings.Contains(text, "\u00A7") {
		// This string uses the old text-based approach to formatting the resulting string. This
		// system is incompatible with inheriting any chat component properties, and therefore
		// has its own system of formatting. The text must be 'scanned' over instead. Using a
		// color formatting code resets all formatting and uses the color code for the proceeding
		// text, while using a decorator formatting code (bold, italics, etc.) will apply that
		// property without resetting the proceeding text. Using '§r' will reset the text, as
		// well as a new line ('\n').

		var (
			tree []Item = make([]Item, 0)
			item Item   = Item{
				Text:       "",
				Decorators: make([]decorators.Decorator, 0),
			}
			r *strings.Reader = strings.NewReader(text)
		)

		// Iterate over the string as long as there still is text to be processed.
		for r.Len() > 0 {
			// Read the next character from the reader.
			char, n, err := r.ReadRune()

			if err != nil {
				return nil, err
			}

			if n < 1 {
				break
			}

			// If the character read is a new-line, push the current item onto the
			// list of items, and set the current one to a new line item with no
			// properties.
			if char == '\n' {
				tree = append(tree, item)

				item = Item{
					Text:       "\n",
					Decorators: make([]decorators.Decorator, 0),
				}

				continue
			}

			// If the current character is also not a formatting code ('§'), add it
			// to the current item text without doing anything else.
			if char != '\u00A7' {
				item.Text += string(char)

				continue
			}

			// Since we now know that formatting is being used, read the next character
			// to check what formatting to apply.
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
					// Parse the color code
					color, ok := colors.Parse(code)

					if !ok {
						break
					}

					// If the current item has no text or decorators, there is no point
					// in doing anything else except just setting the current item to
					// the new color.
					if len(item.Text) == 0 && len(item.Decorators) == 0 {
						item.Color = &color

						break
					}

					// Push the current item into the list of items if there is
					// any text on it, before attempting to apply the color to
					// the remaining text.
					if len(item.Text) > 0 {
						tree = append(tree, item)
					}

					// Reset the current item to use the new color with no extra
					// decorators.
					item = Item{
						Text:       "",
						Color:      &color,
						Decorators: make([]decorators.Decorator, 0),
					}

					break
				}
			case 'k', 'l', 'm', 'n', 'o':
				{
					// Parse the formatting code
					decorator, ok := decorators.Parse(code)

					if !ok {
						break
					}

					// If the current item has no text, there is no point in doing anything
					// else except just pushing the decorator to the item.
					if len(item.Text) == 0 {
						item.Decorators = append(item.Decorators, decorator)

						break
					}

					// Push the current item into the list of items before
					// creating another item that uses this decorator.
					tree = append(tree, item)

					// Reset the current item to use a clone of the current
					// item that has the new decorator.
					item = Item{
						Text:       "",
						Color:      item.Color,
						Decorators: append(item.Decorators, decorator),
					}

					break
				}
			case 'r':
				{
					// Do nothing if there is already no text, color, or decorators.
					if len(item.Text) == 0 && len(item.Decorators) == 0 && item.Color == nil {
						break
					}

					// Append the current item onto the list of items.
					tree = append(tree, item)

					// Reset the current item to have no text, decorators, or color.
					item = Item{
						Text:       "",
						Decorators: make([]decorators.Decorator, 0),
					}

					break
				}
			}
		}

		return append(tree, item), nil
	}

	var (
		color          *colors.Color          = nil
		decoratorsList []decorators.Decorator = make([]decorators.Decorator, 0)
	)

	// Color
	if rawColor, ok := props["color"]; ok {
		parsed, ok := colors.Parse(rawColor)

		if ok {
			color = &parsed
		}
	}

	// Decorators
	for k, v := range decorators.PropertyMap {
		if value, ok := props[k]; ok && parseBool(value) {
			decoratorsList = append(decoratorsList, v)
		}
	}

	return []Item{
		{
			Text:       text,
			Color:      color,
			Decorators: decoratorsList,
		},
	}, nil
}

func parseBool(value interface{}) bool {
	// Although bool is the most common form of a boolean value used in the chat
	// component type, the standard is not properly documented and several types
	// are used out in the wild. We must support them all, even if I have never
	// seen integers used, does not hurt to add them /shrug

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

func mergeChatProperties(parent, child map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range parent {
		// The "extra" property cannot be inherited, or this will cause an infinite loop.
		if k == "extra" {
			continue
		}

		result[k] = v
	}

	for k, v := range child {
		result[k] = v
	}

	return result
}

func cleanupTree(result []Item) []Item {
	newResult := make([]Item, 0)

	for i := 0; i < len(result); i++ {
		j := len(newResult) - 1

		if i != 0 && result[i].IsSameAs(newResult[j]) {
			newResult[j].Text += result[i].Text
		} else {
			newResult = append(newResult, result[i])
		}
	}
	return newResult
}
