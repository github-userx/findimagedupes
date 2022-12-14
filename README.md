findimagedupes [![License](http://img.shields.io/:license-gpl3-blue.svg)](http://www.gnu.org/licenses/gpl-3.0.html) [![Pipeline status](https://gitlab.com/opennota/findimagedupes/badges/master/pipeline.svg)](https://gitlab.com/opennota/findimagedupes/commits/master)
==============

findimagedupes finds visually similar or duplicate images.

# Install

The dependencies:

- Debian or Ubuntu: `apt-get install libmagic-dev libjpeg-dev libpng-dev libtiff5-dev libheif-dev`
- RHEL, CentOS or Fedora: `yum install file-devel libjpeg-devel libpng-devel libtiff-devel libheif-devel`
- Mac OS X:

```
brew install libmagic
brew install libjpeg
brew install libpng
brew install libtiff
brew install dcraw
brew install libheif
```

Then:

    go install gitlab.com/opennota/findimagedupes@latest

# Use

Search for similar images in the `~/Images` directory:

    findimagedupes ~/Images

...and its subdirectories:

    findimagedupes -R ~/Images

The same but use [feh](https://feh.finalrewind.org/) to display the duplicates.

    findimagedupes -R -p feh ~/Images

If no arguments are specified, findimagedupes will print all the available arguments and their default values.

# Donate

**Bitcoin (BTC):** `1PEaahXKwJvNJGJa2PXtPFLNYYigmdLXct`

**Ethereum (ETH):** `0x83e9607E693467Cb344244Df10f66c036eC3Dc53`

# Also

There is a [Perl script](http://www.jhnc.org/findimagedupes/) by that name, which uses a different hashing algorithm.
