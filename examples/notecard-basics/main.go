package main

import (
	"fmt"
	"time"

	"github.com/blues/note-go/notecard"
)

const productUID = "com.your-company.your-project"

var (
	card *notecard.Context
)

func main() {
	var err error
	card, err = notecard.Open(cardInterface, port, speed)
	if err != nil {
		println(err)
		return
	}

	configure()

	for {
		temp, voltage, err := temperatureAndVoltage()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			time.Sleep(1 * time.Second)
			continue
		} else {
			fmt.Printf("Temperature: %f, Voltage: %f\n", temp, voltage)
		}

		err = addNote(temp, voltage)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		fmt.Println("Note added")

		time.Sleep(5 * time.Second)
	}
}

func configure() error {
	req := notecard.Request{Req: notecard.ReqHubSet, ProductUID: productUID, Mode: "continuous"}
	_, err := card.TransactionRequest(req)

	return err
}

func temperatureAndVoltage() (float64, float64, error) {
	rsp, err := card.TransactionRequest(notecard.Request{Req: notecard.ReqCardTemp})
	if err != nil {
		return 0, 0, err
	}

	temp := rsp.Value

	rsp, err = card.TransactionRequest(notecard.Request{Req: notecard.ReqCardVoltage})
	if err != nil {
		return 0, 0, err
	}

	voltage := rsp.Value

	return temp, voltage, nil
}

func addNote(temp float64, voltage float64) error {
	data := map[string]interface{}{"temp": temp, "voltage": voltage}
	_, err := card.TransactionRequest(notecard.Request{Req: notecard.ReqNoteAdd, Body: &data, Sync: true})

	return err
}
