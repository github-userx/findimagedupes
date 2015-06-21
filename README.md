findimagedupes [![License](http://img.shields.io/:license-gpl3-blue.svg)](http://www.gnu.org/licenses/gpl-3.0.html)
==============

findimagedupes finds visually similar or duplicate images.

# Install

You need [pHash](http://www.phash.org/) library installed. Then

    go get github.com/opennota/findimagedupes

# Usage

Search for similar images in the `~/Images` directory and its subdirectories.

    findimagedupes ~/Images

The same but use feh to display the duplicates.

    findimagedupes -v feh ~/Images

If no arguments are specified, findimagedupes will print all the available arguments and their default values.
