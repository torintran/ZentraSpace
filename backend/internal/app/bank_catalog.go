package app

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v5"
)

// Catalog snapshot from https://api.vietqr.io/v2/banks (2026-11-27).
//
//go:embed bank_catalog.json
var bankCatalogJSON string

type BankOption struct {
	BIN, Code, Name string
}

func seedBankCatalog(ctx context.Context, db *pgx.Conn) error {
	_, err := db.Exec(ctx, `
		INSERT INTO bank_catalog (bin, code, name)
		SELECT bin, code, name FROM jsonb_to_recordset($1::jsonb) AS entry(bin text, code text, name text)
		ON CONFLICT (bin) DO UPDATE SET code = EXCLUDED.code, name = EXCLUDED.name
	`, bankCatalogJSON)
	return err
}

func (s *Store) bankOptions(ctx context.Context) ([]BankOption, error) {
	rows, err := s.pool.Query(ctx, "SELECT bin, code, name FROM bank_catalog ORDER BY lower(name), bin")
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[BankOption])
}
