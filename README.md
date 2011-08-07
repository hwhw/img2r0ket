# A simple image converter for the r0ket badge written in Go

This little tool converts PBM bitmaps to a byte stream that can be written to the display buffer in the r0ket badge. To learn more about the r0ket badge, see
[their homepage] [1]. It will work on Linux only for now, I think, since it uses [golzo] [2] for doing (optional) LZO compression. You can use this compression in an [LZO enhanced] [2] r0ket firmware. However, uncompressed output is supported by the basic r0ket firmware as-is.

## Installation

	goinstall github.com/hwhw/img2r0ket

## Usage

`img2r0ket` will read from standard input by default and output to standard output. It will accept the following options:

**-i <file>**	specifies an input file

**-o <file>**	specifies an output file

**-x**		will make img2r0ket invert pixel data

**-L**		will make img2r0ket emit compressed output, prepended by a two-byte, lower endian formatted short number that specifies the number of compressed bytes that follow. If this number is 0, then the block could not be compressed and will follow uncompressed in full length.

## Examples

Create a single uncompressed image

	pngtopnm test.png | pamscale -xyfill 96 68 | ppmtopgm | pnmnorm | pamditherbw | img2r0ket -x > TEST.IMG

This would produce a LZO compressed image stream suitable for "playing" on the r0ket:

	mkdir frames
	ffmpeg -ss '00:00:41' -vframes 850 -i /somevideo.avi -f image2 -r 12.5 -s 112x68 "frames/%08d.pgm"
	ls -1 frames/ | while read f; do \
		cat "frames/$f" | \
		pamcut 8 0 96 68 | \
		pnmnorm | \
		pamditherbw | \
		pamtopnm | \
		img2r0ket -x -L | \
		cat >> LZOTEST.VID;
	done
	rm -rf frames

## Copying/License

The software is free software, licensed under a MIT-style license. See the *LICENSE* file for the terms.


[1]: http://r0ket.badge.events.ccc.de/ "r0ket homepage"
[2]: http://github.com/hwhw/golzo "golzo homepage"
[3]: http://github.com/hwhw/r0ket "my forked r0ket firmware sources with LZO support (in firmware/loadable/anim.c)"
