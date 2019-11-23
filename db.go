// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// This program is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General
// Public License for more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	"gitlab.com/opennota/phash"
)

type entry struct {
	path    string
	fp      uint64
	lastmod int64
}

type DB struct {
	db             *sql.DB
	mu             sync.RWMutex // Protects following.
	preparedGet    *sql.Stmt
	preparedUpsert *sql.Stmt
}

func OpenDatabase(dbpath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS fingerprints (path TEXT PRIMARY KEY, fp INTEGER, lastmod DATETIME)"); err != nil {
		return nil, err
	}

	get, err := db.Prepare("SELECT fp FROM fingerprints WHERE path = ? AND lastmod = ?")
	if err != nil {
		return nil, err
	}

	upsert, err := db.Prepare("INSERT OR REPLACE INTO fingerprints (path, fp, lastmod) VALUES (?, ?, ?)")
	if err != nil {
		return nil, err
	}

	return &DB{
		db:             db,
		preparedGet:    get,
		preparedUpsert: upsert,
	}, nil
}

func (db *DB) Get(ctx context.Context, path string, modtime int64) (uint64, bool, error) {
	var fp int64
	db.mu.RLock()
	row := db.preparedGet.QueryRowContext(ctx, path, modtime)
	err := row.Scan(&fp)
	db.mu.RUnlock()
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}

	return uint64(fp), true, nil
}

func (db *DB) GetAll(ctx context.Context) ([]entry, error) {
	rows, err := db.db.QueryContext(ctx, "SELECT path, fp FROM fingerprints")
	if err != nil {
		return nil, err
	}
	var results []entry
	for rows.Next() {
		var path string
		var fp int64
		if err := rows.Scan(&path, &fp); err != nil {
			return nil, err
		}
		results = append(results, entry{path: path, fp: uint64(fp)})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (db *DB) Upsert(ctx context.Context, path string, modtime int64, fp uint64) error {
	db.mu.Lock()
	_, err := db.preparedUpsert.ExecContext(ctx, path, int64(fp), modtime)
	db.mu.Unlock()
	return err
}

func (db *DB) Prune(ctx context.Context) error {
	rows, err := db.db.QueryContext(ctx, "SELECT path, fp, lastmod FROM fingerprints")
	if err != nil {
		return err
	}
	defer rows.Close()

	var path string
	var fp int64
	var lastmod int64
	var toDelete []string
	var toUpdate []entry
	for rows.Next() {
		if err := rows.Scan(&path, &fp, &lastmod); err != nil {
			if !strings.Contains(err.Error(), "Scan error on column index 2") {
				return err
			}
			fp, lastmod = 0, 0
		}

		fi, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				toDelete = append(toDelete, path)
				continue
			}
			log.Error("ERROR:", err)
			continue
		}

		newlastmod := fi.ModTime().UnixNano()
		if lastmod != newlastmod {
			newfp, err := phash.ImageHashDCT(path)
			if err != nil {
				continue
			}

			toUpdate = append(toUpdate, entry{path, newfp, newlastmod})
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if len(toDelete) == 0 && len(toUpdate) == 0 {
		return nil
	}

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if len(toDelete) > 0 {
		stmt, err := tx.PrepareContext(ctx, "DELETE FROM fingerprints WHERE path = ?")
		if err != nil {
			return err
		}

		for _, path := range toDelete {
			if _, err := stmt.ExecContext(ctx, path); err != nil {
				return err
			}
		}
	}

	if len(toUpdate) > 0 {
		stmt, err := tx.PrepareContext(ctx, "UPDATE fingerprints SET fp = ?, lastmod = ? WHERE path = ?")
		if err != nil {
			return err
		}

		for _, entry := range toUpdate {
			if _, err := stmt.ExecContext(ctx, int64(entry.fp), entry.lastmod, entry.path); err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) Close() error {
	return db.db.Close()
}
