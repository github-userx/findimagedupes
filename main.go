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
	"context"
	"flag"
	"fmt"
	stdlog "log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/rakyll/magicmime"
	"gitlab.com/opennota/phash"
)

var log quietVar

type result struct {
	fp   uint64
	path string
}

func resultWorker(m map[uint64][]string, in <-chan result, done chan struct{}) {
	for r := range in {
		m[r.fp] = append(m[r.fp], r.path)
	}

	close(done)
}

type request struct {
	path    string
	modTime int64
}

func worker(ctx context.Context, db *DB, in <-chan request, out chan<- result, done chan struct{}) {
	defer close(done)

	mm, err := magicmime.NewDecoder(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR)
	if err != nil {
		log.Fatal(err)
	}
	defer mm.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case m, open := <-in:
			if !open {
				return
			}

			var abspath string
			var fp uint64
			haveFP := false

			if db != nil {
				abspath, _ = filepath.Abs(m.path)
				var err error
				fp, haveFP, err = db.Get(ctx, abspath, m.modTime)
				switch {
				case err == context.Canceled:
					return
				case err != nil:
					log.Error("ERROR:", err)
				}
			}

			if !haveFP {
				mimetype, err := mm.TypeByFile(m.path)
				if err != nil {
					log.Warnf("WARNING: %s: %v", m.path, err)
					continue
				}

				if !strings.HasPrefix(mimetype, "image/") {
					continue
				}

				fp, err = phash.ImageHashDCT(m.path)
				if err != nil {
					log.Warnf("WARNING: %s: %v", m.path, err)
					continue
				}

				if db != nil {
					err := db.Upsert(ctx, abspath, m.modTime, fp)
					switch {
					case err == context.Canceled:
						return
					case err != nil:
						log.Error("ERROR:", err)
					}
				}
			}

			res := result{fp: fp, path: m.path}
			select {
			case <-ctx.Done():
				return
			case out <- res:
			}
		}
	}
}

func process(ctx context.Context, depth int, spinner *Spinner, work chan<- request) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		spinner.Spin(path)

		if err != nil {
			log.Warnf("WARNING: %s: %v", path, err)
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

		req := request{
			path:    path,
			modTime: info.ModTime().UnixNano(),
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case work <- req:
		}

		return nil
	}
}

func main() {
	stdlog.SetFlags(0)

	var (
		threshold int
		recurse   bool
		noCompare bool
		program   string
		args      string
		dbPath    string
		prune     bool
		jobs      int
	)

	defaultJobs := runtime.NumCPU()

	flag.IntVar(&threshold, "t", 0, "Hamming distance threshold (0..63)")
	flag.IntVar(&threshold, "threshold", 0, "")

	flag.BoolVar(&recurse, "R", false, "Search for images recursively")
	flag.BoolVar(&recurse, "recurse", false, "")

	flag.BoolVar(&noCompare, "n", false, "Don't look for duplicates")
	flag.BoolVar(&noCompare, "no-compare", false, "")

	flag.StringVar(&program, "p", "", "Launch program (in foreground) to view each set of dupes")
	flag.StringVar(&program, "program", "", "")

	flag.StringVar(&args, "args", "", "Pass additions arguments to the program")

	flag.StringVar(&dbPath, "f", "", "File to use as a fingerprint database")
	flag.StringVar(&dbPath, "fp", "", "")
	flag.StringVar(&dbPath, "db", "", "")
	flag.StringVar(&dbPath, "fingerprints", "", "")

	flag.BoolVar(&prune, "P", false, "Remove fingerprint data for images that do not exist any more")
	flag.BoolVar(&prune, "prune", false, "")

	flag.IntVar(&jobs, "j", defaultJobs, "Number of jobs to use for image processing")
	flag.IntVar(&jobs, "jobs", defaultJobs, "")

	flag.Var(&log, "q", "Quiet mode (no warnings, if given once; no errors either, if given twice)")
	flag.Var(&log, "quiet", "")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: findimagedupes [options] [file...]

    Options:
       -t, --threshold=AMOUNT         Use AMOUNT as threshold of similarity (0..63; default 0)
       -R, --recurse                  Search recursively for images inside subdirectories
       -n, --no-compare               Don't look for duplicates
       -p, --program=PROGRAM          Launch PROGRAM (in foreground) to view each set of dupes
           --args=ARGUMENTS           Pass additional ARGUMENTS to the program before the filenames;
                                          e.g. for feh, '-. -^ "%%u / %%l - %%wx%%h - %%n"'
       -f, --fingerprints=FILE        Use FILE as fingerprint database
       -P, --prune                    Remove fingerprint data for images that do not exist any more
       -j, --jobs                     Number of jobs to use for image processing (default %d)
       -q, --quiet                    If this option is given, warnings are not displayed; if it is
                                          given twice, non-fatal errors are not displayed either

       -h, --help                     Show this help

`, defaultJobs)
	}
	flag.Parse()

	if prune && dbPath == "" {
		log.Fatal("--prune used without -f")
	}

	if args != "" && program == "" {
		log.Fatal("--args used without --program")
	}

	if noCompare && program != "" {
		log.Fatal("--no-compare used with --program")
	}

	if noCompare && dbPath == "" {
		log.Fatal("--no-compare is useless without -f")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		select {
		case <-ctx.Done():
		case <-sig:
			// Interrupt will not immediately exit the program,
			// instead we signal to stop processing new data and
			// allow the program to exit cleanly.
			cancel()
		}
	}()

	var db *DB
	if dbPath != "" {
		var err error
		db, err = OpenDatabase(dbPath)
		if err != nil {
			log.Fatal(err)
		}

		if prune {
			if err := db.Prune(ctx); err != nil {
				db.Close()
				if err == context.Canceled {
					os.Exit(1)
				}
				log.Fatal(err)
			}
		}
	}

	if flag.NArg() == 0 {
		if prune {
			os.Exit(0)
		}
		flag.Usage()
		os.Exit(1)
	}

	programArgs := parseArgs(args)

	spinner := NewSpinner()

	// Search for image files and compute hashes.
	maxDepth := 1
	if recurse {
		maxDepth = -1
	}
	m := make(map[uint64][]string)

	results := make(chan result)

	workC := make(chan request)
	workDone := make(chan chan struct{}, jobs)
	for i := 0; i < jobs; i++ {
		done := make(chan struct{})
		go worker(ctx, db, workC, results, done)
		workDone <- done
	}
	close(workDone)

	resultDone := make(chan struct{})
	go resultWorker(m, results, resultDone)

	for _, d := range flag.Args() {
		walkFn := process(ctx, maxDepth, spinner, workC)
		filepath.Walk(d, walkFn)
	}

	close(workC)
	for done := range workDone {
		<-done
	}
	close(results)
	<-resultDone

	if db != nil {
		err := db.Close()
		if err != nil {
			log.Errorf("Error closing DB: %v", err)
		}
	}

	// Exit immediately if the program was interrupted.
	select {
	case <-ctx.Done():
		os.Exit(1)
	default:
	}

	signal.Stop(sig) // Stop handling interrupts gracefully.
	cancel()         // Cleanup.

	spinner.Stop()

	if noCompare {
		os.Exit(0)
	}

	// Produce repeatable output.
	hashes := make([]uint64, 0, len(m))
	for h, files := range m {
		hashes = append(hashes, h)
		sort.Strings(files)
	}
	sort.Slice(hashes, func(i, j int) bool { return hashes[i] < hashes[j] })

	// Find similar hashes.
	if threshold > 0 {
		for i := 0; i < len(hashes)-1; i++ {
			for j := i + 1; j < len(hashes); j++ {
				h1 := hashes[i]
				h2 := hashes[j]

				d := phash.HammingDistance(h1, h2)
				if d <= threshold {
					m[h1] = append(m[h1], m[h2]...)
					delete(m, h2)
				}
			}
		}
	}

	// Print or view duplicates.
	for _, h := range hashes {
		files := m[h]
		if len(files) < 2 {
			continue
		}

		sort.Strings(files)
		if program == "" {
			fmt.Println(strings.Join(files, " "))
		} else {
			args := append(programArgs, files...)
			cmd := exec.Command(program, args...)
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				log.Errorf("ERROR: %s %s: %v", program, strings.Join(args, " "), err)
			}
		}
	}
}
