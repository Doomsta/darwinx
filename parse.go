package darwinx

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"math"
	"strconv"
	"strings"
)

var (
	ErrInvalidSegment   = errors.New("invalid migration segment")
	ErrInvalidHeader    = errors.New("invalid header (expected: '<type> <version> <description>')")
	ErrInvalidVersion   = errors.New("invalid version (float64 required)")
	ErrEmptyDescription = errors.New("empty description")
	ErrEmptyScript      = errors.New("empty script")
)

func MigrationsFromString(content string) ([]Migration, error) {
	return MigrationsFromReader(strings.NewReader(content))
}

func MigrationsFromReader(r io.Reader) ([]Migration, error) {
	sc := bufio.NewScanner(r)

	const maxLine = 2 * 1024 * 1024
	buf := make([]byte, 64*1024)
	sc.Buffer(buf, maxLine)

	var (
		migrations        []Migration
		header            string
		sqlLines          []string
		inSegment         bool
		seenNonEmptySince bool
		segIndex          int
	)

	flush := func() error {
		if strings.TrimSpace(header) == "" && len(sqlLines) == 0 {
			return nil
		}
		hz := strings.TrimSpace(header)
		if hz == "" {
			return fmt.Errorf("%w: header missing (segment %d)", ErrInvalidSegment, segIndex)
		}

		fields := strings.Fields(hz)
		if len(fields) < 3 {
			return fmt.Errorf("%w: got %q (segment %d)", ErrInvalidHeader, hz, segIndex)
		}

		verStr := fields[1]
		desc := strings.TrimSpace(strings.TrimPrefix(hz, fields[0]+" "+verStr))
		if desc == "" {
			return fmt.Errorf("%w (segment %d)", ErrEmptyDescription, segIndex)
		}

		v, err := strconv.ParseFloat(verStr, 64)
		if err != nil {
			return fmt.Errorf("%w: %v (segment %d)", ErrInvalidVersion, err, segIndex)
		}
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("%w: NaN/Inf disallowed (segment %d)", ErrInvalidVersion, segIndex)
		}
		if v == 0 && math.Signbit(v) {
			v = 0
		}

		script := strings.TrimSpace(strings.Join(sqlLines, "\n"))
		if script == "" {
			return fmt.Errorf("%w (segment %d)", ErrEmptyScript, segIndex)
		}

		migrations = append(migrations, Migration{
			Version:     v,
			Description: desc,
			Script:      script,
		})

		header = ""
		sqlLines = sqlLines[:0]
		inSegment = false
		seenNonEmptySince = false
		segIndex++
		return nil
	}

	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")

		if isSeparatorLine(line) {
			if err := flush(); err != nil {
				return nil, err
			}
			continue
		}

		if inSegment && looksLikeHeader(line) {
			if err := flush(); err != nil {
				return nil, err
			}
		}

		if !inSegment {
			if strings.TrimSpace(line) == "" {
				continue
			}
			header = strings.TrimSpace(line)
			inSegment = true
			seenNonEmptySince = false
			continue
		}

		if strings.TrimSpace(line) == "" && !seenNonEmptySince {
			continue
		}
		if strings.TrimSpace(line) != "" {
			seenNonEmptySince = true
		}
		sqlLines = append(sqlLines, line)
	}

	if err := sc.Err(); err != nil {
		return nil, err
	}
	if err := flush(); err != nil {
		return nil, err
	}
	return migrations, nil
}

func isSeparatorLine(s string) bool {
	t := strings.TrimSpace(s)
	if len(t) < 4 {
		return false
	}
	for i := 0; i < len(t); i++ {
		if t[i] != '-' {
			return false
		}
	}
	return true
}

func looksLikeHeader(line string) bool {
	// Strikt: Nur echte Header-Zeilen, die mit "---- " beginnen
	s := strings.TrimLeft(line, " \t")
	return strings.HasPrefix(s, "---- ")
}
