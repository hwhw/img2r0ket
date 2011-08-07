/*
* Program to convert PBM images to r0ket (http://r0ket.badge.events.ccc.de/) format images
*
* (c) 2011 Hans-Werner Hilse
*
* this software is published under the know-what-you're-doing licence:
*
* 1. you may use this software for whatever you want, given that you know what you're doing.
* 2. author of this software isn't responsible for anything since you knew what you're doing.
* 3. if you have still questions, chances are you do not fullfill requirements
*----------------------------------------------------------------------------------------------
*
* Changelog:
*
* 0.1: accepts raw PBM images
*/
package main

import (
	"flag"
	"log"
	"os"
	"io"
	"io/ioutil"
	"strconv"

	"golzo"
)

var inputFile = flag.String("i", "/proc/self/fd/0", "input file (PBM format)")
var outputFile = flag.String("o", "/proc/self/fd/1", "output file (r0ket format)")
var invert = flag.Bool("x", false, "invert output bits")
var compress = flag.Bool("L", false, "compress with LZO, outputs 16bits of compressed block length + compressed block")

const (
	/* software will enforce these dimensions: */

	IMG_WIDTH = 96
	IMG_HEIGHT = 68
)

func readImg(imgReader io.Reader) (r0ketimg []byte) {
	input, err := ioutil.ReadAll(imgReader) // I'm lazy
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("got %d bytes of data", len(input))

	// check header:
	if input[0] != 'P' || input[1] != '4' || (input[2] >= '0' && input[2] <= '9') {
		log.Fatal("is not a raw PBM image")
	}

	// scan for beginning of a number:
	p1 := 3
	for input[p1] < '0' || input[p1] > '9' {
		if input[p1] == '#' {
			// comment line, ignore until next EOL
			for input[p1] != '\n' {
				p1++ // TODO: check length!
			}
		} else {
			p1++
		}
	}
	// number starts at <p1>
	p2 := p1
	for input[p2] >= '0' && input[p2] <= '9' {
		p2++
	}
	width, _ := strconv.Atoi(string(input[p1:p2]))

	p1 = p2
	for input[p1] < '0' || input[p1] > '9' {
		if input[p1] == '#' {
			for input[p1] != '\n' {
				p1++
			}
		} else {
			p1++
		}
	}
	p2 = p1
	for input[p2] >= '0' && input[p2] <= '9' {
		p2++
	}
	height, _ := strconv.Atoi(string(input[p1:p2]))

	if width != IMG_WIDTH || height != IMG_HEIGHT {
		log.Fatal("Error, wrong size")
	}

	p2++

	if len(input) < (p2 + ((width + 7) >> 3)*height) {
		log.Fatal("Not enough data in the image, image file broken?")
	}

	bitmap := make([]byte, width * height) // TODO: optimize this out into a single parse?
	stride := (width + 7) >> 3
	log.Printf("width: %d, stride: %d", width, stride)

	/* parse input data into bitmap */
	for y := 0; y < height; y++ {
		for x := 0; x < stride; x++ {
			for b := 0; b < 8; b++ {
				tx := (x+1)*8 - 1 - b
				if tx < width {
					bitmap[y*width + tx] = input[p2] & 1
				}
				input[p2] >>= 1
			}
			p2++
		}
	}

	r0ketimg = make([]byte, width * ((height+7) >> 3))
	/* create output data from bitmap: */
	for y:=0; y < (height+7) >> 3; y++ {
		for x:=0; x < width; x++ {
			for b:=uint(0); b < 8 ; b++ {
				ox := ((width - x)) - 1
				oy := ((height - y*8) - 1) - int(b)
				px := ox + oy * width
				if px >= 0 {
					if (bitmap[px] > 0 && !*invert) || (bitmap[px] == 0 && *invert) {
						r0ketimg[y*width + x] |= 1 << b
					}
				}
			}
		}
	}
	return
}

func main() {
	flag.Parse()

	f, err := os.Open(*inputFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fout, err := os.OpenFile(*outputFile, os.O_WRONLY | os.O_TRUNC | os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()

	out := readImg(f)
	if *compress {
		ret := golzo.Lzo_init()
		if ret != 0 {
			log.Fatal("LZO error: ", ret)
		}
		comp := golzo.Lzo1x_999_compress(out)
		if len(comp) > len(out) {
			log.Printf("not compressable: %d bytes of output for %d bytes input after trying", len(comp), len(out))
			fout.Write([]byte{0,0})
		} else {
			log.Printf("compressed: %d bytes of output for %d bytes input", len(comp), len(out))
			fout.Write([]byte{byte(len(comp) & 0xff), byte((len(comp) >> 8) & 0xff)})
			out = comp
		}
	}
	fout.Write(out)
}
