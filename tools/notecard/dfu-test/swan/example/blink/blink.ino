// This example generates a .bin that you can fish out of Arduino's temp folders to
// generate the blink tests herein.  If you'd like to control where Arduino puts
// the bin files,
// 1. Use Arduino's preferences to find the text
//    "More preferences can be edited directly in the file" and open/edit that file.
// 2. Within that file find a line that looks like this:
//       build.path=/Users/rozzie/tmp
// 3. Change that path to be any path that you choose, and when you build
//    you'll find that Arduino uses GCC commands to place output there.

// Control loop delay
#define LOOP_DELAY	150

void setup()
{

	// Delay in order to give Arduino IDE enough time to switch serial port from DFU to debug
	delay(2500);

	// LED
	pinMode(LED_BUILTIN, OUTPUT);

	// Debug init
	Serial.begin(115200);
	Serial.printf("=============================================================\n");

}

void loop()
{
	static int count = 0;
	Serial.printf("iterations: %d\n", count++);
	int on = (count & 1) != 0 ? HIGH : LOW;
	digitalWrite(LED_BUILTIN, on);
	delay(LOOP_DELAY);
}
