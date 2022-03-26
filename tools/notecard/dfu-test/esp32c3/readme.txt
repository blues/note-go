The BIN files herein, and the layout herein, will work for any Arduino sketch built for
the esp32c3. Before building it, you must select the partitioning on the Arduino
Tools menu, Partition Scheme option, and select:
    No OTA (1MB APP, 3MB FATFS)

This will build the app in a way such that it is compatible with the layout in the
files in this directory, and how the binpack will be laid out.

Then, build your app and supply your app's bin to:
    package.sh your-arduino-app.bin

...and a .binpack file will be generated in the same folder as your bin.
