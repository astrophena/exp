// Package tspgstore implements a ipn.StateStore that persists the data in a
// PostgreSQL database.
package tspgstore

import (
	"database/sql"

	_ "github.com/lib/pq"
	"tailscale.com/ipn"
)

// Making sure that we're adhering to the ipn.StateStore interface.
var _ ipn.StateStore = (*Store)(nil)

// New returns a new ipn.StateStore that persists the data in a PostgreSQL
// database.
func New(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	s := &Store{db: db}

	const sql = `
CREATE TABLE IF NOT EXISTS tailscale_state (
	key varchar(400) not null primary key,
	data bytea not null
);
`

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	if _, err := tx.Exec(sql); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s, nil
}

// Store is a ipn.StateStore that persists the data in a PostgreSQL database.
type Store struct {
	db *sql.DB
}

// Close closes the database connection.
func (s *Store) Close() error { return s.db.Close() }

// ReadState implements the ipn.StateStore interface.
func (s *Store) ReadState(id ipn.StateKey) ([]byte, error) {
	query := `SELECT data::bytea FROM tailscale_state WHERE key = $1`
	var bs []byte
	err := s.db.QueryRow(query, id).Scan(&bs)
	if err == sql.ErrNoRows {
		return nil, ipn.ErrStateNotExist
	}
	return bs, err
}

// WriteState implements the ipn.StateStore interface.
func (s *Store) WriteState(id ipn.StateKey, bs []byte) error {
	query := `INSERT INTO tailscale_state (key, data) VALUES ($1, $2::bytea)
	               ON CONFLICT (key) DO UPDATE SET data = $2::bytea`
	_, err := s.db.Exec(query, id, bs)
	if err != nil {
		return err
	}
	return nil
}
