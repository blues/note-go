#!/bin/bash

# Copyright 2018 Inca Roads LLC.  All rights reserved. 
# Use of this source code is governed by licenses granted by the
# copyright holder including that found in the LICENSE file.

# If this resin.io env var is asserted, halt so we can play around in SSH
while [[ $HALT != "" ]]
do
      echo "HALT asserted to enable debugging..."
      sleep 10s
done

# Use wiringpi to set GPIO38 to allow RPI3 to go to 1.2A on USB rather than 600mA
gpio -g mode 38 out
gpio -g write 38 1

# Load the i2c-dev kernel module
# Each i2c bus can address 127 independent i2c devices, and most
# linux systems contain several buses.
modprobe i2c-dev

# Turn of ModemManager (courtesy of petrosagg) so that it doesn't send garbage
# to the modem port
set -o errexit
export DBUS_SYSTEM_BUS_ADDRESS=unix:path=/host/run/dbus/system_bus_socket
dbus-send \
  --system \
  --print-reply \
  --dest=org.freedesktop.systemd1 \
  /org/freedesktop/systemd1 \
  org.freedesktop.systemd1.Manager.StopUnit 'string:ModemManager.service' 'string:replace'

# Run this in the foreground so we can watch the log
# NOTE that you must set the ARGS variable to do the test that you want
while [[ true ]]
do
    echo "*** run.sh STARTING APP"
    $GOPATH/bin/testcard
    echo "*** run.sh APPLICATION EXITED"
    sleep 15s
done
