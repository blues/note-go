In order to utilize this feature, the Notecard must be enabled for "direct DFU" by
wiring AUX3, AUX4, and RX/TX to the host MCU.
AUX3 -> boot pin
AUX4 -> reset pin
RX -> host RX pin (not tx)
TX -> host TX pin (not rx)

For the Nucleo-WL55JC1 board, use this command:
{"req":"dfu.direct","name":"stm32"}

In wiring it, you can refer to this:
https://www.st.com/resource/en/user_manual/dm00622917-stm32wl-nucleo64-board-mb1389-stmicroelectronics.pdf
...however there are MANY ERRORS in how it describes pins/signals/sockets.

What is most important is:
0 GND
1 AUX3
2 AUX4
3 RX
4 TX


LEFTMOST CONNECTOR (CN7)
x x
x x
x x
1 x

Left Arduino connectors:
X
X
2
X
X
X
0
X

X
X
3
X
X
X

Right Arduino connectors:
X
X
X
X
X
X
X
X
4
X

X
X
X
X
X
X
X
X
