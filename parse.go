package darwinx

import (
	"fmt"
	"strconv"
	"strings"
)

func MigrationsFromString(content string) (migrations []Migration, err error) {
	content = strings.Trim(content, "\n")
	for _, s := range strings.Split(content, "----") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		parts := strings.SplitN(s, "\n", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("migration block must have a header line and SQL body; got %q", s)
		}
		header := strings.TrimSpace(parts[0])
		sql := strings.Trim(parts[1], "\n ")

		headerParts := strings.SplitN(header, " ", 2)
		if len(headerParts) < 2 {
			return nil, fmt.Errorf("invalid migration header (need version and description): %q", header)
		}

		version, err := strconv.ParseFloat(headerParts[0], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid header version part: %w", err)
		}

		migrations = append(migrations, Migration{
			Version:     version,
			Description: strings.TrimSpace(headerParts[1]),
			Script:      sql,
		})
	}

	return migrations, nil
}
