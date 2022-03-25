The BIN files herein, and the layout herein, will work for any Arduino sketch built for
a generic ESP32 board. The images and scripts herein are built for a 4MB version, using:
Tools menu, Board:
    ESP32 Arduino / ESP32 Dev Module
Tools menu, Flash Size:
    4MB (32Mb)
Tools menu, Partition Scheme option, and select:
    Default 4MB with ffat (1.2MB APP/1.5MB FATFS)

This will build the app in a way such that it is compatible with the layout in the
files in this directory, and how the binpack will be laid out.

Then, build your app and supply your app's bin to:
    package.sh your-arduino-app.bin

...and a .binpack file will be generated in the same folder as your bin.

--------------------------

In order to utilize this feature, the Notecard must be enabled for "direct DFU" by
wiring AUX3, AUX4, and RX/TX to the host MCU.
AUX3 -> boot pin
AUX4 -> reset pin
RX -> host RX pin (not tx)
TX -> host TX pin (not rx)

{"req":"dfu.direct","name":"esp32"}
