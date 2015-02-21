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
	quiet      = flag.Bool("q", false, "Quiet mode (no warnings)")
	threshold  = flag.Int("t", 0, "Hamming distance threshold")
	viewer     = flag.String("v", "", `Image viewer, e.g. -v feh; if no viewer is specified (default), findimagedupes will print similar files to the standard output`)
	viewerArgs StringSlice

	mm, _ = magicmime.New(magicmime.MAGIC_MIME_TYPE | magicmime.MAGIC_SYMLINK | magicmime.MAGIC_ERROR)
	hmap  = make(map[uint64][]string)
)

// ProcessFile computes a hash of the file if it is an image file,
// and saves it in the hmap map.
func ProcessFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		if !*quiet {
			log.Printf("WARNING: %s: %v", path, err)
		}

		return nil
	}

	if !info.Mode().IsRegular() {
		return nil
	}

	mimetype, err := mm.TypeByFile(path)
	if err != nil {
		if !*quiet {
			log.Printf("WARNING: %s: %v", path, err)
		}

		return nil
	}

	if !strings.HasPrefix(mimetype, "image/") {
		return nil
	}

	hash, err := phash.ImageHashDCT(path)
	if err != nil {
		if !*quiet {
			log.Printf("WARNING: %s: %v", path, err)
		}

		return nil
	}

	if hash == 0 {
		if !*quiet {
			log.Printf("WARNING: %s: cannot compute hash", path)
		}

		return nil
	}

	hmap[hash] = append(hmap[hash], path)

	return nil
}

func main() {
	log.SetFlags(0)

	flag.Var(&viewerArgs, "a", `A viewer argument; you can specify multiple arguments, e.g. -a -. -a -^ -a "[%u / %l] [%wx%h] %n"`)
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: findimagedupes [options] directory [directory...]")
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	// Search for image files and compute hashes.
	for _, d := range flag.Args() {
		filepath.Walk(d, ProcessFile)
	}

	// Find similar images.
	hashes := make([]uint64, 0, len(hmap))
	for h := range hmap {
		hashes = append(hashes, h)
	}
	for i := 0; i < len(hashes)-1; i++ {
		for j := i + 1; j < len(hashes); j++ {
			h1 := hashes[i]
			h2 := hashes[j]

			d := phash.HammingDistance(h1, h2)
			if d <= *threshold {
				hmap[h1] = append(hmap[h1], hmap[h2]...)
				delete(hmap, h2)
			}
		}
	}

	// Print or view duplicates.
	for _, files := range hmap {
		if len(files) < 2 {
			continue
		}

		if *viewer == "" {
			fmt.Println(strings.Join(files, " "))
		} else {
			args := append(viewerArgs, files...)
			cmd := exec.Command(*viewer, args...)
			err := cmd.Run()
			if err != nil {
				log.Printf("ERROR: %s %s: %v", *viewer, strings.Join(args, " "), err)
			}
		}
	}
}
