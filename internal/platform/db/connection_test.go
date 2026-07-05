package db

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/isyll/go-grpc-starter/pkg/config"
)

func TestBuildDSNQuotesSpecialCharacters(t *testing.T) {
	creds := config.DBCredentials{
		Host:     "db.internal",
		Port:     "5432",
		User:     "app user",
		Password: `p@ss 'word' \with\ specials`,
		DBName:   "appdb",
		SSLMode:  "require",
	}

	cfg, err := pgxpool.ParseConfig(BuildDSN(creds))
	require.NoError(t, err)
	assert.Equal(t, "db.internal", cfg.ConnConfig.Host)
	assert.Equal(t, uint16(5432), cfg.ConnConfig.Port)
	assert.Equal(t, "app user", cfg.ConnConfig.User)
	assert.Equal(t, `p@ss 'word' \with\ specials`, cfg.ConnConfig.Password)
	assert.Equal(t, "appdb", cfg.ConnConfig.Database)
}

func TestBuildDSNAppendsExtraParams(t *testing.T) {
	creds := config.DBCredentials{
		Host: "localhost", Port: "5432", User: "u", Password: "p",
		DBName: "d", SSLMode: "disable",
	}
	cfg, err := pgxpool.ParseConfig(
		BuildDSN(creds, "search_path='public,auth'"),
	)
	require.NoError(t, err)
	assert.Equal(t, "public,auth", cfg.ConnConfig.RuntimeParams["search_path"])
}
