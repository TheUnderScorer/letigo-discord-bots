package markdownutil

import "fmt"

// Link formats a string as a Markdown hyperlink using the provided link and label.
func Link(link string, label string) string {
	return fmt.Sprintf("[%s](%s)", label, link)
}
