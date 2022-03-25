In order to utilize this feature, the Notecard must be enabled for "direct DFU" by
wiring AUX3, AUX4, and RX/TX to the host MCU.
AUX3 -> boot pin (which is a challenge on V1/V2 swans)
AUX4 -> reset pin
RX -> host RX pin (not tx)
TX -> host TX pin (not rx)

For the swan, use this command:
{"req":"dfu.direct","name":"stm32"}

