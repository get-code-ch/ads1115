package ads1115

import (
	"log"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/host"
)

type Ads1115 struct {
	Device      *i2c.Dev
	Name        string `json:"name"`
	Address     int    `json:"address"`
	Description string `json:"description"`
}

const (
	conversionReg = 0x00
	configReg     = 0x01
	lowThreshReg  = 0x02
	highThreshReg = 0x03

	// Configuration Register MSB[7] OS field
	OsField = 7
	OsWNoEffect = 0b0
	OsWStartSingleConversion = 0b1 //reset
	OsRPerformingConversion = 0b0
	OsRConversionReady = 0b1

	// Configuration Register MSB[6:4] MUX field
	MuxField = 4
	MuxA0A1 = 0b000 //reset
	MuxA0A3 = 0b001
	MuxA1A3 = 0b010
	MuxA2A3 = 0b011
	MuxA0GND = 0b100
	MuxA1GND= 0b101
	MuxA2GND = 0b110
	MuxA3GND = 0b111

	// Configuration Register MSB[3:1] PGA (Gain field)
	PGAField = 1
	PGA6144 = 0b000
	PGA6144Scale = 0.1875 / 1000
	PGA4096 = 0b001
	PGA4096Scale = 0.125 / 1000
	PGA2048 = 0b010 //reset
	PGA2048Scale = 0.0625 / 1000
	PGA1024 = 0b011
	PGA1024Scale = 0.03125 / 1000
	PGA512 = 0b100
	PGA512Scale = 0.015625 / 1000
	PGA256 = 0b101
	PGA256Scale = 0.0078125 / 1000

	// Configuration Register MSB[0] Mode
	ModeField = 0
	ModeContinuous = 0
	ModeSingle = 1 //reset

	// Configuration Register LSB[7:5] Data Rate
	DRField = 7
	SPS8 = 0b000
	SPS16 = 0b001
	SPS32 = 0b010
	SPS64 = 0b011
	SPS128 = 0b100 //reset
	SPS250 = 0b101
	SPS475 = 0b110
	SPS86 = 0b111

	// Configuration Register LSB[4] Comp Mode
	CMField = 4
	CMTraditional = 0b0 //reset
	CMWindowComparator = 0b1

	// Configuration Register LSB[3] Comp polarity
	CPField = 3
	CPActiveLow = 0b0 //reset
	CPActiveHigh = 0b1

	// Configuration Register LSB[2] Latching comparator
	LCField = 2
	LCNon = 0b0 //reset
	LCYes = 0b1

	// Configuration Register LSB[1:0] Comparator queue
	QField = 0
	QOne = 0b00
	QTwo = 0b01
	QFour = 0b10
	QDisable = 0b11 //reset
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

	ConfMsb := byte(0)
	ConfLsb := byte(0)

	ConfMsb = OsWStartSingleConversion << OsField | MuxA3GND << MuxField | PGA4096 << PGAField | ModeContinuous << ModeField
	log.Printf("ConfMsb %b\n", ConfMsb)
	ConfLsb = SPS16 << DRField | CMTraditional << CMField | CPActiveLow << CPField | LCNon << LCField | QDisable << QField
	module.Device.Write([]byte{configReg,ConfMsb,ConfLsb})

	return err
}

func ReadConversionRegister(module *Ads1115) (int16, float64, []byte) {
	regValue := []byte{0,0}

	value := int16(0)

	module.Device.Tx([]byte{conversionReg},regValue)
	value = 0 | (int16(regValue[0]) << 8) | int16(regValue[1])

	return value, float64(value) * PGA4096Scale, regValue
}
