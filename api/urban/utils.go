package urban

import (
	"net/url"
	"strings"
)

func ParseUrlArgs(urlArgs string) (string, string) {
	term, _ := url.QueryUnescape(urlArgs)
	term = strings.TrimSpace(term)

	atUser := ""
	if len(term) > 0 {
		if term[0] == '!' {
			termSplitted := strings.Split(term, " ")
			termSplitted = termSplitted[1:]
			term = strings.Join(termSplitted, " ")
		}

		if strings.Contains(term, "@") {
			termSplitted := strings.Split(term, " ")
			term = ""

			for _, word := range termSplitted {
				if word[0] == '@' {
					atUser = word + " "
					continue
				}

				term = term + word + " "
			}
		}
	}

	term = strings.TrimSpace(term)
	atUser = strings.TrimSpace(atUser)

	return term, atUser
}
