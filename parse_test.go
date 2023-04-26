package darwinx

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
