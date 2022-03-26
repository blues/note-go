// This example generates a .bin that you can fish out of Arduino's temp folders to
// generate the blink tests herein.  If you'd like to control where Arduino puts
// the bin files,
// 1. Use Arduino's preferences to find the text
//    "More preferences can be edited directly in the file" and open/edit that file.
// 2. Within that file find a line that looks like this:
//       build.path=/Users/rozzie/tmp
// 3. Change that path to be any path that you choose, and when you build
//    you'll find that Arduino uses GCC commands to place output there.

// Control LED flashing speed
#define LED_DELAY	1000

// ESP32
#include "esp_partition.h"
#include "esp_system.h"
#include "esp_ota_ops.h"
#include "esp_app_format.h"
#include "esp_flash_partitions.h"
#include "nvs.h"
#include "nvs_flash.h"
#include "SPIFFS.h"
#include "FFat.h"

// Assume that this board has a neopixel
#include <Adafruit_NeoPixel.h>
#define PIN        8
#define NUMPIXELS  1
Adafruit_NeoPixel pixels(NUMPIXELS, PIN, NEO_GRB + NEO_KHZ800);

void setup()
{

	// Delay in order to give Arduino IDE enough time to switch serial port from DFU to debug
	delay(2500);

	// Debug init
	Serial.begin(115200);

	// Show ESP32 info
	Serial.printf("\n");
	Serial.printf("=============================================================\n");
	esp_chip_info_t info;
	esp_chip_info(&info);
	switch (info.model) {
	case CHIP_ESP32:
		Serial.printf("ESP32\n");
		break;
	case CHIP_ESP32S2:
		Serial.printf("ESP32-S2\n");
		break;
	case CHIP_ESP32S3:
		Serial.printf("ESP32-S3\n");
		break;
	case CHIP_ESP32C3:
		Serial.printf("ESP32-C3\n");
		break;
	case CHIP_ESP32H2:
		Serial.printf("ESP32-H2\n");
		break;
	default:
		Serial.printf("ESP32 TYPE UNKNOWN (model:%d)\n", info.model);
		break;
	}
	Serial.printf("  Revision: %d\n", info.revision);
	Serial.printf("  Cores: %d\n", info.cores);
	if ((info.features & CHIP_FEATURE_EMB_FLASH) != 0) {
		Serial.printf("  Embedded Flash\n");
	} else {
		Serial.printf("  NO Embedded Flash\n");
	}
	if ((info.features & CHIP_FEATURE_WIFI_BGN) != 0) {
		Serial.printf("  2.4GHz WiFi\n");
	} else {
		Serial.printf("  NO WiFi\n");
	}
	if ((info.features & CHIP_FEATURE_BT) != 0) {
		Serial.printf("  Bluetooth Classic\n");
	} else {
		Serial.printf("  NO Bluetooth Classic\n");
	}
	if ((info.features & CHIP_FEATURE_IEEE802154) != 0) {
		Serial.printf("  IEEE 802.155.4\n");
	} else {
		Serial.printf("  NO IEEE 802.155.4\n");
	}
	Serial.printf("-------------------------------------------------------------\n");
    const esp_partition_t *partition;

	// Display flash partitions
	showPartitions();

	// Show NVS key/value pairs
	showNVS();

	// Show SPIFF file system if present
	showSPIFFS();

	// Show FatFs file system if present
	showFATFS();

	// Done showing info about the ESP32
	Serial.printf("=============================================================\n");

	// Initialize neopixel
	pixels.begin();
	pixels.clear();

}

void loop()
{
	setPixelColor(255, 0, 0);	// Red
	delay(LED_DELAY);
	setPixelColor(0, 255, 0);	// Green
	delay(LED_DELAY);
	setPixelColor(0, 0, 255);	// Blue
	delay(LED_DELAY);
	setPixelColor(0, 255, 255);	// Cyan
	delay(LED_DELAY);
	setPixelColor(255, 0, 255);	// Magenta
	delay(LED_DELAY);
	setPixelColor(255, 255, 0);	// Amber
	delay(LED_DELAY);
	setPixelColor(255, 255, 255);// White
	delay(LED_DELAY);
	setPixelColor(0, 0, 0);		// Black
	delay(LED_DELAY);
}

// Set the pixel color
void setPixelColor(uint8_t r, uint8_t g, uint8_t b)
{
    pixels.setPixelColor(0, pixels.Color(r, g, b));
    pixels.show();
}

// Get the string name of type enum values used in this example
const char* get_type_str(esp_partition_type_t type)
{
    switch(type) {
	case ESP_PARTITION_TYPE_APP:
		return "ESP_PARTITION_TYPE_APP";
	case ESP_PARTITION_TYPE_DATA:
		return "ESP_PARTITION_TYPE_DATA";
	default:
		return "UNKNOWN_PARTITION_TYPE";
    }
}

// Get the string name of subtype enum values used in this example
const char* get_subtype_str(esp_partition_subtype_t subtype)
{
    switch(subtype) {
	case ESP_PARTITION_SUBTYPE_DATA_NVS:
		return "ESP_PARTITION_SUBTYPE_DATA_NVS";
	case ESP_PARTITION_SUBTYPE_DATA_PHY:
		return "ESP_PARTITION_SUBTYPE_DATA_PHY";
	case ESP_PARTITION_SUBTYPE_APP_FACTORY:
		return "ESP_PARTITION_SUBTYPE_APP_FACTORY";
	case ESP_PARTITION_SUBTYPE_DATA_FAT:
		return "ESP_PARTITION_SUBTYPE_DATA_FAT";
	default:
		return "UNKNOWN_PARTITION_SUBTYPE";
    }
}

// print OTA info as it relates to the partition
void showPartitionOTAInfo(const char *prefix, const esp_partition_t *part)
{
	static char attributes[256];
	strcpy(attributes, prefix);
	if (esp_ota_get_boot_partition() == part) {
		strcat(attributes, "BOOT ");
	}
	if (esp_ota_get_running_partition() == part) {
		strcat(attributes, "RUNNING ");
	}
	if (esp_ota_get_running_partition() == part) {
		strcat(attributes, "NEXT ");
	}
	esp_app_desc_t desc;
	if (esp_ota_get_partition_description(part, &desc) == ESP_OK) {
		if (desc.project_name[0] != '\0') {
			strcat(attributes, desc.project_name);
			strcat(attributes, " ");
		}
		if (desc.version[0] != '\0') {
			strcat(attributes, "ver:'");
			strcat(attributes, desc.version);
			strcat(attributes, "' ");
		}
		if (desc.date[0] != '\0') {
			strcat(attributes, desc.date);
			strcat(attributes, " ");
		}
		if (desc.time[0] != '\0') {
			strcat(attributes, desc.time);
			strcat(attributes, " ");
		}
		if (desc.idf_ver[0] != '\0') {
			strcat(attributes, "IDF:'");
			strcat(attributes, desc.idf_ver);
			strcat(attributes, "' ");
		}
	}
	if (attributes[strlen(prefix)] == '\0') {
		strcat(attributes, "(empty)");
	}
	Serial.printf("%s\n", attributes);
}

// Show the partitions
void showPartitions()
{
    esp_partition_iterator_t it;

	Serial.printf("Partitions:\n");

    Serial.printf("  App:\n");
    it = esp_partition_find(ESP_PARTITION_TYPE_APP, ESP_PARTITION_SUBTYPE_ANY, NULL);
    for (; it != NULL; it = esp_partition_next(it)) {
        const esp_partition_t *part = esp_partition_get(it);
        Serial.printf("    '%s' offset:0x%x length:%d %s %s\n",
					  part->label, part->address, part->size,
					  get_type_str(part->type), get_subtype_str(part->subtype));
		showPartitionOTAInfo("        ", part);
    }
    esp_partition_iterator_release(it);

    Serial.printf("  Data:\n");
    it = esp_partition_find(ESP_PARTITION_TYPE_DATA, ESP_PARTITION_SUBTYPE_ANY, NULL);
    for (; it != NULL; it = esp_partition_next(it)) {
        const esp_partition_t *part = esp_partition_get(it);
        Serial.printf("    '%s' offset:0x%x length:%d %s %s\n",
					  part->label, part->address, part->size,
					  get_type_str(part->type), get_subtype_str(part->subtype));
    }
    esp_partition_iterator_release(it);

}

// Get NVS entry type name
const char *nvsTypeName(nvs_type_t type)
{
	switch (type) {
	case NVS_TYPE_U8:
		return "uint8_t";
	case NVS_TYPE_I8:
		return "int8_t";
	case NVS_TYPE_U16:
		return "uint16_t";
	case NVS_TYPE_I16:
		return "int16_t";
	case NVS_TYPE_U32:
		return "uint32_t";
	case NVS_TYPE_I32:
		return "int32_t";
	case NVS_TYPE_U64:
		return "uint64_t";
	case NVS_TYPE_I64:
		return "int64_t";
	case NVS_TYPE_STR:
		return "string";
	case NVS_TYPE_BLOB:
		return "blob";
	}
	return "UNKNOWN";
}

// Show NVS
void showNVS()
{
	Serial.printf("NVS partition:\n");
    nvs_iterator_t it = nvs_entry_find("nvs", NULL, NVS_TYPE_ANY);
	if (it == NULL) {
		Serial.printf("  (empty)\n");
	}
    while (it != NULL) {
	    nvs_entry_info_t info;
        nvs_entry_info(it, &info);
        Serial.printf("%s %s (%s)\n", info.namespace_name, info.key, nvsTypeName(info.type));
        it = nvs_entry_next(it);
    }
}

// Show SPI Flash File System
void showSPIFFS()
{
	Serial.printf("SPI Flash File System:\n");
	if (!SPIFFS.begin(true)) {
		Serial.printf("can't load SPIFFS library\n");
		return;
	}
	File root = SPIFFS.open("/");
	if (!root) {
		Serial.printf("can't open /\n");
		return;
	}
	File file = root.openNextFile();
	if (!file) {
		Serial.printf("  (empty)\n");
	} else {
		while (file) {
			Serial.printf("  %s\n", file.name());
			file = root.openNextFile();
		}
	}
}

// Show FATFS File System
void showFATFS()
{
	Serial.printf("FAT File System:\n");
	if (!FFat.begin(true)) {
		Serial.printf("can't load FFat library\n");
		return;
	}
	File root = FFat.open("/");
	if (!root) {
		Serial.printf("can't open /\n");
		return;
	}
	File file = root.openNextFile();
	if (!file) {
		Serial.printf("  (empty)\n");
	} else {
		while (file) {
			Serial.printf("  %s\n", file.name());
			file = root.openNextFile();
		}
	}
}
