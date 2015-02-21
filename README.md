findimagedupes
==============

findimagedupes finds visually similar or duplicate images.

# Install

    go get github.com/opennota/findimagedupes

# Usage

Search for similar images in the `~/Images` directory and its subdirectories.

    findimagedupes ~/Images

The same but use feh to display the duplicates.

    findimagedupes -v feh ~/Images

If no arguments are specified, findimagedupes will print all the available arguments and their default values.

# License

GNU GPL v3+

