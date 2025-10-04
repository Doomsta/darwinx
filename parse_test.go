package darwinx

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func Test_MigrationsFromString_Simple(t *testing.T) {
	input := `---- 0.3 creating table config
CREATE TABLE config
(
    id    bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    key   VARCHAR(255),
    value VARCHAR(255)
);
`
	expected := []Migration{
		{
			Version:     0.3,
			Description: `creating table config`,
			Script: `CREATE TABLE config
(
    id    bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    key   VARCHAR(255),
    value VARCHAR(255)
);`,
		},
	}
	actual, err := MigrationsFromString(input)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func Test_MigrationsFromReader_IgnoresDashesInsideSQL(t *testing.T) {
	input := `
---- 3.0 section with dashes in SQL
-- this is a comment ---- should not split
INSERT INTO t (c) VALUES ('---- inside string');
`
	got, err := MigrationsFromString(input)
	require.NoError(t, err)
	require.Len(t, got, 1)
	expectedScript := `-- this is a comment ---- should not split
INSERT INTO t (c) VALUES ('---- inside string');`
	assert.Equal(t, expectedScript, got[0].Script)
}

func Test_MigrationsFromReader_EmptyScript(t *testing.T) {
	input := `
---- 7.0 desc
  
  
`
	_, err := MigrationsFromString(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty script")
}

func Test_MigrationsFromReader_InvalidFloat_NaN_Inf(t *testing.T) {
	cases := []string{
		"---- NaN bad\nSELECT 1;",
		"---- Inf bad\nSELECT 1;",
		"---- -Inf bad\nSELECT 1;",
	}
	for _, in := range cases {
		_, err := MigrationsFromString(in)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid version")
	}
}

func Test_MigrationsFromReader_MinusZeroNormalized(t *testing.T) {
	input := "---- -0 desc\nSELECT 1;"
	got, err := MigrationsFromString(input)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 0.0, got[0].Version)
}

func Test_MigrationsFromReader_Multiple_WithDo(t *testing.T) {
	input := `---- 1 create
CREATE TABLE demo (id BIGSERIAL PRIMARY KEY, message text);
---- 2 do-rename
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema = 'public' AND table_name = 'demo' AND column_name = 'message'
  ) THEN
    ALTER TABLE public.demo RENAME COLUMN message TO note;
  END IF;
END
$$;
---- 3 index
CREATE INDEX idx_demo_note ON public.demo (note);`

	got, err := MigrationsFromString(input)
	require.NoError(t, err)
	require.Len(t, got, 3)

	assert.Equal(t, 1.0, got[0].Version)
	assert.Equal(t, 2.0, got[1].Version)
	assert.Equal(t, 3.0, got[2].Version)
}

func Test_MigrationsFromString_Trim(t *testing.T) {
	input := `---- 0.3 creating table config
    
CREATE TABLE config
(
    id    bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY
);
   
`
	expected := []Migration{
		{
			Version:     0.3,
			Description: `creating table config`,
			Script: `CREATE TABLE config
(
    id    bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY
);`,
		},
	}
	actual, err := MigrationsFromString(input)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func Test_MigrationsFromString_Multi(t *testing.T) {
	input := `
---- 0.1 desc1
CREATE TABLE config1 (id int PRIMARY KEY);
---- 0.2 desc2
CREATE TABLE config2 (id int PRIMARY KEY);
---- 0.3 desc3
CREATE TABLE config3 (id int PRIMARY KEY);
---- 0.4 desc4
CREATE TABLE config4 (id int PRIMARY KEY);
`
	expected := []Migration{
		{
			Version:     0.1,
			Description: `desc1`,
			Script:      `CREATE TABLE config1 (id int PRIMARY KEY);`,
		},
		{
			Version:     0.2,
			Description: `desc2`,
			Script:      `CREATE TABLE config2 (id int PRIMARY KEY);`,
		},
		{
			Version:     0.3,
			Description: `desc3`,
			Script:      `CREATE TABLE config3 (id int PRIMARY KEY);`,
		},
		{
			Version:     0.4,
			Description: `desc4`,
			Script:      `CREATE TABLE config4 (id int PRIMARY KEY);`,
		},
	}
	actual, err := MigrationsFromString(input)
	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}
func Test_MigrationsFromReader_CRLF(t *testing.T) {
	input := "---- 5.0 windows line endings\r\n\r\nCREATE TABLE x (\r\n  id bigint\r\n);\r\n"
	got, err := MigrationsFromReader(strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, got, 1)
	expected := "CREATE TABLE x (\n  id bigint\n);"
	assert.Equal(t, expected, got[0].Script)
}

func Test_MigrationsFromReader_InvalidHeader_MissingParts(t *testing.T) {
	input := `
---- onlyversion
SELECT 1;
`
	_, err := MigrationsFromReader(strings.NewReader(input))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid header")
}

func Test_MigrationsFromReader_TrailingWhitespaceAndBlankAfterScript(t *testing.T) {
	input := `---- 9.1 trailing
SELECT 1;
   
  
`
	got, err := MigrationsFromString(input)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "SELECT 1;", got[0].Script)
}
