package darwinx

import (
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

func MigrationsFromString(content string) (migrations []Migration, err error) {
	content = strings.Trim(content, "\n")
	for _, s := range strings.Split(content, "----") {
		if s == "" {
			continue
		}
		parts := strings.SplitN(s, "\n", 2)
		if len(parts) != 2 {
			return nil, nil
		}
		header := parts[0]
		sql := parts[1]

		headerParts := strings.SplitN(header, " ", 3)
		if len(parts) != 2 {
			return nil, errors.New("invalid header parts")
		}

		version, err := strconv.ParseFloat(headerParts[1], 64)
		if err != nil {
			return nil, errors.Wrap(err, "invalid header version part")
		}

		migrations = append(migrations, Migration{
			Version:     version,
			Description: strings.Trim(headerParts[2], " "),
			Script:      strings.Trim(sql, "\n "),
		})
	}

	return migrations, nil
}
