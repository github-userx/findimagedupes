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
	"database/sql"
	"flag"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/opennota/phash"
)

type entry struct {
	path    string
	fp      uint64
	lastmod string
}

var (
	db             *sql.DB
	preparedGet    *sql.Stmt
	preparedUpsert *sql.Stmt
)

func init() {
	flag.BoolVar(&prune, "prune", false, "Remove fingerprint data for images that do not exist any more")

	dbpath := os.ExpandEnv("$HOME/.findimagedupes.sqlite")
	var err error
	db, err = sql.Open("sqlite3", dbpath)
	if err != nil {
		log.Fatalf("sql.Open: %v", err)
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS fingerprints (path TEXT PRIMARY KEY, fp INTEGER, lastmod DATETIME)")
	if err != nil {
		log.Fatal(err)
	}

	preparedGet, err = db.Prepare("SELECT fp FROM fingerprints WHERE path = ? AND lastmod = ?")
	if err != nil {
		log.Fatal(err)
	}

	preparedUpsert, err = db.Prepare("INSERT OR REPLACE INTO fingerprints (path, fp, lastmod) VALUES (?, ?, ?)")
	if err != nil {
		log.Fatal(err)
	}
}

func iso8601(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func Get(path string, modtime time.Time) (uint64, bool) {
	lastmod := iso8601(modtime)
	row := preparedGet.QueryRow(path, lastmod)
	var fp int64
	err := row.Scan(&fp)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false
		}
		if !*quiet {
			log.Print(err)
		}
		return 0, false
	}

	return uint64(fp), true
}

func Upsert(path string, modtime time.Time, fp uint64) {
	lastmod := iso8601(modtime)
	_, err := preparedUpsert.Exec(path, int64(fp), lastmod)
	if err != nil {
		if !*quiet {
			log.Print(err)
		}
	}
}

func Prune() {
	rows, err := db.Query("SELECT path, fp, lastmod FROM fingerprints")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var path string
	var fp int64
	var tm time.Time
	var toDelete []string
	var toUpdate []entry
	for rows.Next() {
		err := rows.Scan(&path, &fp, &tm)
		if err != nil {
			log.Fatal(err)
		}

		fi, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				toDelete = append(toDelete, path)
				continue
			}
			log.Printf("os.Stat: %v", err)
			continue
		}

		lastmod := iso8601(tm)
		newlastmod := iso8601(fi.ModTime())
		if lastmod != newlastmod {
			newfp, err := phash.ImageHashDCT(path)
			if err != nil {
				continue
			}

			if newfp == 0 {
				continue
			}

			toUpdate = append(toUpdate, entry{path, newfp, newlastmod})
		}
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	if len(toDelete) == 0 && len(toUpdate) == 0 {
		return
	}

	tx, err := db.Begin()

	if len(toDelete) > 0 {
		stmt, err := tx.Prepare("DELETE FROM fingerprints WHERE path = ?")
		if err != nil {
			log.Fatal(err)
		}

		for _, path := range toDelete {
			_, err := stmt.Exec(path)
			if err != nil {
				log.Print(err)
			}
		}
	}

	if len(toUpdate) > 0 {
		stmt, err := tx.Prepare("UPDATE fingerprints SET fp = ?, lastmod = ? WHERE path = ?")
		if err != nil {
			log.Fatal(err)
		}

		for _, entry := range toUpdate {
			_, err := stmt.Exec(int64(entry.fp), entry.lastmod, entry.path)
			if err != nil {
				log.Print(err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Print(err)
	}
}
