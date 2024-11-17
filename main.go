package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pierreyves258/psu/psu"
)

func main() {
	var tty, file, delimiter string
	var cutoff float64
	flag.StringVar(&tty, "p", "/dev/ttyUSB1", "Serial port")
	flag.StringVar(&file, "o", "test.csv", "CSV file")
	flag.StringVar(&delimiter, "d", ",", "CSV delimiter")
	flag.Float64Var(&cutoff, "c", 0.032, "Charge CutOff current")
	flag.Parse()

	dc310s, err := psu.NewPSU(tty)
	if err != nil {
		log.Fatal(err)
	}
	defer dc310s.Destroy()

	f, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString("time" + delimiter + "voltage" + delimiter + "current\n")
	if err != nil {
		log.Fatal(err)
	}

	err = dc310s.SetData(psu.SetCurrent, 1.15)
	if err != nil {
		log.Fatal(err)
	}

	err = dc310s.SetData(psu.SetVoltage, 8.4)
	if err != nil {
		log.Fatal(err)
	}

	err = dc310s.SetData(psu.SetOutput, true)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second) // Wait for current to flow

	for {
		current, err := dc310s.GetData(psu.GetCurrent)
		if err != nil {
			continue
		}
		voltage, err := dc310s.GetData(psu.GetVoltage)
		if err != nil {
			continue
		}

		str := fmt.Sprintf("%s%s%f%s%f\n", time.Now().Format("2006-01-02 15:04:05"), delimiter, voltage, delimiter, current)

		_, err = f.WriteString(str)
		if err != nil {
			fmt.Println("ERROR WRITE", err)
			// No blocker
		}

		if current.(float64) < cutoff {
			break
		}
		time.Sleep(1 * time.Second)
	}

	err = dc310s.SetData(psu.SetOutput, false)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)
}
