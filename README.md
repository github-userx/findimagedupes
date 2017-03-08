findimagedupes [![License](http://img.shields.io/:license-gpl3-blue.svg)](http://www.gnu.org/licenses/gpl-3.0.html)
==============

findimagedupes finds visually similar or duplicate images.

# Installation

findimagedupes uses [rakyll/magicmime](https://github.com/rakyll/magicmime),
which requires `libmagic`. Install it as follows:

- Debian or Ubuntu: `apt-get install libmagic-dev`
- RHEL, CentOS or Fedora: `yum install file-devel`
- Mac OS X: `brew install libmagic`

Then

    go get github.com/opennota/findimagedupes

# Usage

Search for similar images in the `~/Images` directory:

    findimagedupes ~/Images

...and its subdirectories:

    findimagedupes -R ~/Images

The same but use [feh](https://feh.finalrewind.org/) to display the duplicates.

    findimagedupes -R -p feh ~/Images

If no arguments are specified, findimagedupes will print all the available arguments and their default values.

# Also

There is a [Perl script](http://www.jhnc.org/findimagedupes/) by that name, which uses a different hashing algorithm.
