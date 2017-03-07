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

// Find visually similar or duplicate images.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rakyll/magicmime"

	"github.com/opennota/phash"
)

var (
	threshold   int
	recurse     bool
	noCompare   bool
	program     string
	programArgs string
	dbPath      string
	prune       bool
	quiet       bool

	db *DB

	hmap = make(map[uint64][]string)

	spinner *Spinner
)

func init() {
	err := magicmime.Open(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR)
	if err != nil {
		log.Fatal(err)
	}
}

func process(depth int) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if !quiet {
				log.Printf("WARNING: %s: %v", path, err)
			}

			return nil
		}

		if !info.Mode().IsRegular() {
			if info.Mode().IsDir() {
				if depth == 0 {
					return filepath.SkipDir
				}
				if depth > 0 {
					depth--
				}
			}
			return nil
		}

		var abspath string
		var fp uint64
		haveFP := false

		if db != nil {
			abspath, _ = filepath.Abs(path)
			var err error
			fp, haveFP, err = db.Get(abspath, info.ModTime())
			if err != nil {
				log.Print("ERROR:", err)
			}
		}

		if haveFP {
			spinner.Spin(path)
		} else {
			mimetype, err := magicmime.TypeByFile(path)
			if err != nil {
				if !quiet {
					log.Printf("WARNING: %s: %v", path, err)
				}

				return nil
			}

			if !strings.HasPrefix(mimetype, "image/") {
				return nil
			}

			spinner.Spin(path)

			fp, err = phash.ImageHashDCT(path)
			if err != nil {
				if !quiet {
					log.Printf("WARNING: %s: %v", path, err)
				}

				return nil
			}

			if fp == 0 {
				if !quiet {
					log.Printf("WARNING: %s: cannot compute fingerprint", path)
				}

				return nil
			}

			if db != nil {
				if err := db.Upsert(abspath, info.ModTime(), fp); err != nil {
					log.Println("ERROR:", err)
				}
			}
		}

		hmap[fp] = append(hmap[fp], path)

		return nil
	}
}

func main() {
	log.SetFlags(0)

	flag.IntVar(&threshold, "t", 0, "Hamming distance threshold (0..64)")
	flag.IntVar(&threshold, "threshold", 0, "")

	flag.BoolVar(&recurse, "R", false, "Search for images recursively")
	flag.BoolVar(&recurse, "recurse", false, "")

	flag.BoolVar(&noCompare, "n", false, "Don't look for duplicates")
	flag.BoolVar(&noCompare, "no-compare", false, "")

	flag.StringVar(&program, "p", "", "Launch program (in foreground) to view each set of dupes")
	flag.StringVar(&program, "program", "", "")

	flag.StringVar(&programArgs, "args", "", "Pass additions arguments to the program")

	flag.StringVar(&dbPath, "f", "", "File to use as a fingerprint database")
	flag.StringVar(&dbPath, "fp", "", "")
	flag.StringVar(&dbPath, "db", "", "")
	flag.StringVar(&dbPath, "fingerprints", "", "")

	flag.BoolVar(&prune, "P", false, "Remove fingerprint data for images that do not exist any more")
	flag.BoolVar(&prune, "prune", false, "")

	flag.BoolVar(&quiet, "q", false, "Quiet mode (no warnings)")
	flag.BoolVar(&quiet, "quiet", false, "")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, `Usage: findimagedupes [options] [file...]

    Options:
       -t, --threshold=AMOUNT         Use AMOUNT as threshold of similarity
       -R, --recurse                  Search recursively for images inside subdirectories
       -n, --no-compare               Don't look for duplicates
       -p, --program=PROGRAM          Launch PROGRAM (in foreground) to view each set of dupes
           --args=ARGUMENTS           Pass additional ARGUMENTS to the program before the filenames
       -f, --fingerprints=FILE        Use FILE as fingerprint database
       -P, --prune                    Remove fingerprint data for images that do not exist any more
       -q, --quiet                    Quiet mode (no warnings)

       -h, --help
`)
	}
	flag.Parse()

	if prune && dbPath == "" {
		log.Fatal("--prune used without -f")
	}

	if programArgs != "" && program == "" {
		log.Fatal("--args used without --program")
	}

	if noCompare && program != "" {
		log.Fatal("--no-compare used with --program")
	}

	if noCompare && dbPath == "" {
		log.Fatal("--no-compare is useless without -f")
	}

	var err error
	if db, err = OpenDatabase(dbPath); err != nil {
		log.Fatal(err)
	}

	if prune {
		if err := db.Prune(); err != nil {
			log.Fatal(err)
		}
	}

	if flag.NArg() == 0 {
		if prune {
			os.Exit(0)
		}
		flag.Usage()
		os.Exit(1)
	}

	programArgs := parseArgs(programArgs)

	spinner = NewSpinner()

	// Search for image files and compute hashes.
	depth := 1
	if recurse {
		depth = -1
	}
	process := process(depth)
	for _, d := range flag.Args() {
		filepath.Walk(d, process)
	}

	spinner.Stop()

	if noCompare {
		os.Exit(0)
	}

	// Find similar hashes.
	if threshold > 0 {
		hashes := make([]uint64, 0, len(hmap))
		for h := range hmap {
			hashes = append(hashes, h)
		}
		for i := 0; i < len(hashes)-1; i++ {
			for j := i + 1; j < len(hashes); j++ {
				h1 := hashes[i]
				h2 := hashes[j]

				d := phash.HammingDistance(h1, h2)
				if d <= threshold {
					hmap[h1] = append(hmap[h1], hmap[h2]...)
					delete(hmap, h2)
				}
			}
		}
	}

	// Print or view duplicates.
	for _, files := range hmap {
		if len(files) < 2 {
			continue
		}

		if program == "" {
			fmt.Println(strings.Join(files, " "))
		} else {
			args := append(programArgs, files...)
			cmd := exec.Command(program, args...)
			err := cmd.Run()
			if err != nil {
				log.Printf("ERROR: %s %s: %v", program, strings.Join(args, " "), err)
			}
		}
	}
}
