package psu

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
)

type PSU struct {
	port serial.Port
}

type IDF struct { // KIPRIM,<model, eg.DC310S>,<serial no.>,FV:Vx.x.x
	Brand   string
	Model   string
	Serial  string
	Version string
}

const (
	GetID              = "*IDN?"
	GetVoltage         = "MEAS:VOLT?"
	GetCurrent         = "MEAS:CURR?"
	GetPower           = "MEAS:POW?"
	GetSetpointVoltage = "VOLT?"
	GetSetpointCurrent = "CURR?"
	GetLimitVoltage    = "VOLT:LIM?"
	GetLimitCurrent    = "CURR:LIM?"
	SetVoltage         = "VOLT %.3f"
	SetCurrent         = "CURR %.3f"
	SetLimitVoltage    = "VOLT:LIM %.3f"
	SetLimitCurrent    = "CURR:LIM %.3f"
	GetOutput          = "OUTP?"
	SetOutput          = "OUTP %s"
)

var outputStr = map[bool]string{
	true:  "ON",
	false: "OFF",
}

func NewPSU(tty string) (*PSU, error) {
	mode := &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		Parity:   serial.NoParity,
		StopBits: serial.OneStopBit,
	}
	port, err := serial.Open(tty, mode)
	if err != nil {
		return nil, err
	}
	port.SetReadTimeout(3 * time.Second)

	return &PSU{
		port: port,
	}, nil
}

func (psu *PSU) Destroy() {
	psu.port.Close()
	psu = nil
}

func (psu *PSU) SetData(function string, value interface{}) error {
	if function == SetOutput {
		if vBool, ok := value.(bool); ok {
			return psu.sendCommand(function, outputStr[vBool])
		}
		return fmt.Errorf("invalid value %v for bool", value)
	}
	return psu.sendCommand(function, value)
}

func (psu *PSU) sendCommand(function string, value interface{}) error {
	if psu == nil {
		return fmt.Errorf("PSU not initialised")
	}

	psu.port.ResetInputBuffer() // Flush read buffer
	var str string
	if value != nil {
		str = fmt.Sprintf("%s\n", fmt.Sprintf(function, value))
	} else {
		str = fmt.Sprintf("%s\n", function)
	}
	_, err := psu.port.Write([]byte(str))
	if err != nil {
		return fmt.Errorf("error write command [%s][%v]", function, value)
	}

	time.Sleep(500 * time.Millisecond) // Documentation says so...

	return nil
}

func (psu *PSU) readReply() (string, error) {
	res := ""
	for {
		buff := make([]byte, 32)
		n, err := psu.port.Read(buff)
		if err != nil {
			return "", err
		}
		if n == 0 {
			return "", fmt.Errorf("EOF")
		}
		res += string(buff[:n])
		if buff[n-1] == '\n' { // Response end with /r/n
			break
		}
	}

	if len(res) < 2 {
		return "", fmt.Errorf("invalid response size [%s]", res)
	}

	res = res[:len(res)-2] // Remove /r/n

	if res == "ERR" {
		return "", fmt.Errorf("ERR")
	}

	return res, nil
}

func (psu *PSU) GetData(function string) (interface{}, error) {
	if psu == nil {
		return nil, fmt.Errorf("psu not initialised")
	}

	err := psu.sendCommand(function, nil)
	if err != nil {
		return nil, err
	}

	res, err := psu.readReply()
	if err != nil {
		return nil, err
	}

	switch function {
	case GetID:
		spl := strings.Split(res, ",")
		if len(spl) != 4 {
			return nil, fmt.Errorf("invalid id")
		}
		if len(spl[3]) < 4 {
			return nil, fmt.Errorf("invalid version")
		}
		return IDF{
			Brand:   spl[0],
			Model:   spl[1],
			Serial:  spl[2],
			Version: spl[3][4:],
		}, nil
	case GetVoltage, GetCurrent, GetPower, GetSetpointCurrent, GetSetpointVoltage, GetLimitCurrent, GetLimitVoltage:
		return strconv.ParseFloat(res, 64)
	case GetOutput:
		fmt.Println(res)
		return res == "ON", nil
	}

	return nil, err
}
