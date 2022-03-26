#!/bin/bash
if [ -z "$1" ]
then
	echo "Usage to create application packaged with bootloader::"
	echo "    package foo.bin"
	exit 1
fi
../../notecard -output $1 -binpack esp32 0x1000:bootloader.bin 0x8000:partitions.bin 0xe000:boot_app0.bin 0x10000:$1
