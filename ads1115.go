package ads1115

import (
	"log"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/host/v3"
	"strings"
	"time"
)

type Ads1115 struct {
	Device      *i2c.Dev
	Name        string `json:"name"`
	Address     int    `json:"address"`
	Description string `json:"description"`
}

type OS byte
type Mux byte
type PGA byte
type Mode byte
type DR byte
type CM byte
type CP byte
type LC byte
type Queue byte

type Configuration struct {
	OS                 OS    `json:"os"`
	Mux                Mux   `json:"mux"`
	PGA                PGA   `json:"pga"`
	Mode               Mode  `json:"mode"`
	DataRate           DR    `json:"dr"`
	ComparatorMode     CM    `json:"cm"`
	ComparatorPolarity CP    `json:"cp"`
	LatchingComparator LC    `json:"lc"`
	ComparatorQueue    Queue `json:"queue"`
}

const (
	// Register address
	conversionReg = 0x00
	configReg     = 0x01
	lowThreshReg  = 0x02
	highThreshReg = 0x03

	// Configuration Register MSB[7] OS field
	OsField                     = 7
	OsWNoEffect              OS = 0b0 // W
	OsWStartSingleConversion OS = 0b1 // W / reset
	OsRPerformingConversion  OS = 0b0 // R
	OsRConversionReady       OS = 0b1 // R

	// Configuration Register MSB[6:4] MUX field
	MuxField Mux = 4
	MuxA0A1  Mux = 0b000 //reset
	MuxA0A3  Mux = 0b001
	MuxA1A3  Mux = 0b010
	MuxA2A3  Mux = 0b011
	MuxA0GND Mux = 0b100
	MuxA1GND Mux = 0b101
	MuxA2GND Mux = 0b110
	MuxA3GND Mux = 0b111

	// Configuration Register MSB[3:1] PGA (Gain field)
	PGAField         = 1
	PGA6144      PGA = 0b000
	PGA6144Scale     = 187.5 / 1000000
	PGA4096      PGA = 0b001
	PGA4096Scale     = 125 / 1000000
	PGA2048      PGA = 0b010 //reset
	PGA2048Scale     = 62.5 / 1000000
	PGA1024      PGA = 0b011
	PGA1024Scale     = 31.25 / 1000000
	PGA512       PGA = 0b100
	PGA512Scale      = 15.625 / 1000000
	PGA256       PGA = 0b101
	PGA256Scale      = 7.8125 / 1000000

	// Configuration Register MSB[0] Mode
	ModeField           = 0
	ModeContinuous Mode = 0
	ModeSingle     Mode = 1 //reset

	// Configuration Register LSB[7:5] Data Rate
	DRField    = 5
	SPS8    DR = 0b000
	SPS16   DR = 0b001
	SPS32   DR = 0b010
	SPS64   DR = 0b011
	SPS128  DR = 0b100 //reset
	SPS250  DR = 0b101
	SPS475  DR = 0b110
	SPS860  DR = 0b111

	// Configuration Register LSB[4] Comp Mode
	CMField               = 4
	CMTraditional      CM = 0b0 //reset
	CMWindowComparator CM = 0b1

	// Configuration Register LSB[3] Comp polarity
	CPField         = 3
	CPActiveLow  CP = 0b0 //reset
	CPActiveHigh CP = 0b1

	// Configuration Register LSB[2] Latching comparator
	LCField    = 2
	LCNon   LC = 0b0 //reset
	LCYes   LC = 0b1

	// Configuration Register LSB[1:0] Comparator queue
	QField         = 0
	QOne     Queue = 0b00
	QTwo     Queue = 0b01
	QFour    Queue = 0b10
	QDisable Queue = 0b11 //reset
)

func New(device string, name string, address int, description string) (Ads1115, error) {
	log.Printf("I2C Module %s initialization...\n", name)
	var err error
	module := Ads1115{nil, name, address, description}
	if device != "" {
		err = Init(device, module.Address, &module)
	}
	return module, err
}

// Init function initialize MCP28003 after boot or restart of device
func Init(device string, add int, module *Ads1115) error {

	var err error
	var b i2c.Bus

	// I2C Bus initialization
	host.Init()
	if b, err = i2creg.Open(device); err != nil {
		module.Device = nil
		return err
	}
	module.Device = &i2c.Dev{Addr: uint16(add), Bus: b}
	/*
		ConfMsb := byte(0)
		ConfLsb := byte(0)

		ConfMsb = byte(OsWStartSingleConversion)<<OsField | byte(MuxA3GND)<<MuxField | byte(PGA4096<<PGAField) | byte(ModeContinuous<<ModeField)
		log.Printf("ConfMsb %b\n", ConfMsb)
		ConfLsb = byte(SPS16<<DRField) | byte(CMTraditional<<CMField) | byte(CPActiveLow<<CPField) | byte(LCNon<<LCField) | byte(QDisable<<QField)
		module.Device.Write([]byte{configReg, ConfMsb, ConfLsb})

	*/
	return nil
}

func ReadConversionRegister(module *Ads1115, input string) float64 {

	// Config input value to read
	ConfMsb := byte(0)
	ConfLsb := byte(0)

	mux := MuxA0GND
	switch strings.ToUpper(input) {
	case "AIN0":
		mux = MuxA0GND
		break
	case "AIN1":
		mux = MuxA1GND
		break
	case "AIN2":
		mux = MuxA2GND
		break
	case "AIN3":
		mux = MuxA3GND
		break
	}

	ConfMsb = byte(0)<<OsField | byte(mux)<<MuxField | byte(PGA6144<<PGAField) | byte(ModeContinuous<<ModeField)
	ConfLsb = byte(SPS16<<DRField) | byte(CMTraditional<<CMField) | byte(CPActiveLow<<CPField) | byte(LCNon<<LCField) | byte(QOne<<QField)
	module.Device.Write([]byte{configReg, ConfMsb, ConfLsb})

	time.Sleep(250 * time.Millisecond)
	// Reading the value
	regValue := []byte{0, 0}
	value := int16(0)
	module.Device.Tx([]byte{conversionReg}, regValue)
	value = 0 | (int16(regValue[0]) << 8) | int16(regValue[1])
	/*
		log.Printf("end point %s\n", input)
		log.Printf("value %d\n", value)
		log.Printf("float value %.4f\n", float64(value) * PGA6144Scale)
		log.Printf("regvalue %d\n", regValue)
	*/
	//return float64(value) * PGA4096Scale
	return float64(value) * PGA6144Scale
}
