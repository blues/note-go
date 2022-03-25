#!/bin/bash
if [ -z "$1" ]
then
	echo "Usage to create application packaged with bootloader::"
	echo "    package foo.bin"
	exit 1
fi
../../notecard -output $1 -binpack esp32c3 0x0,0x8000:esp-bootloader.bin 0x8000,0x1000:esp-partitions.bin 0xe000,0x2000:esp-otadata.bin 0x10000:$1
