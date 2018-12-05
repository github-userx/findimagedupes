findimagedupes [![License](http://img.shields.io/:license-gpl3-blue.svg)](http://www.gnu.org/licenses/gpl-3.0.html) [![Pipeline status](https://gitlab.com/opennota/findimagedupes/badges/master/pipeline.svg)](https://gitlab.com/opennota/findimagedupes/commits/master)
==============

findimagedupes finds visually similar or duplicate images.

# Install

You'll need Go 1.8 or newer.

The dependencies:

- Debian or Ubuntu: `apt-get install libmagic-dev libjpeg-dev libpng-dev libtiff5-dev`
- RHEL, CentOS or Fedora: `yum install file-devel libjpeg-devel libpng-devel libtiff-devel`
- Mac OS X:

```
brew install libmagic
brew install libjpeg
brew install libpng
brew install libtiff
```

Then:

    go get -u gitlab.com/opennota/findimagedupes

# Use

Search for similar images in the `~/Images` directory:

    findimagedupes ~/Images

...and its subdirectories:

    findimagedupes -R ~/Images

The same but use [feh](https://feh.finalrewind.org/) to display the duplicates.

    findimagedupes -R -p feh ~/Images

If no arguments are specified, findimagedupes will print all the available arguments and their default values.

# Also

There is a [Perl script](http://www.jhnc.org/findimagedupes/) by that name, which uses a different hashing algorithm.
