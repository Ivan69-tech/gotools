package modbus2

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/simonvetter/modbus"
)

type Conf struct {
	Name      []string
	Address   []int
	Size_data []string
	Bit       []int
	Type_data []string
}

type Res struct {
	Name []string
	Res  []int
}

func CreateModbusClient(adresse string, port string) *modbus.ModbusClient {

	client, err := modbus.NewClient(&modbus.ClientConfiguration{
		URL:     "tcp://" + adresse + ":" + port,
		Timeout: 1 * time.Second,
	})

	if err != nil {
		fmt.Printf("failed to create modbus client: %v\n", err)
		os.Exit(1)
	}

	err = client.Open()
	if err != nil {
		fmt.Printf("failed to connect: %v\n", err)
		os.Exit(2)
	}

	return client
}

func (c *Conf) Read(mc *modbus.ModbusClient, r *Res) {

	r.Res = make([]int, len(c.Address))

	for i, j := range c.Address {
		dataSize := c.Size_data[i]
		dataType := c.Type_data[i]
		dataTypeRead := modbus.HOLDING_REGISTER

		switch dataType {
		case "input":
			dataTypeRead = modbus.INPUT_REGISTER
		case "holding":
			dataTypeRead = modbus.HOLDING_REGISTER
		case "coil":
			regs, err := mc.ReadCoil(uint16(j))
			if err != nil {
				fmt.Printf("failed to read coil registers %d: %v\n", j, err)
			}
			if regs == true {
				r.Res[i] = 1
			} else {
				r.Res[i] = 0
			}
		default:
			fmt.Println("can not read register type")
			os.Exit(2)
		}

		switch dataSize {

		case "int16", "uint16":
			regs, err := mc.ReadRegisters(uint16(j), 1, dataTypeRead)
			if err != nil {
				fmt.Printf("failed to read registers %d: %v\n", j, err)
			}
			fmt.Println(regs)
			r.Res[i] = int(regs[0])

		case "int32", "uint32":
			regs, err := mc.ReadRegisters(uint16(j), 2, dataTypeRead)
			if err != nil {
				fmt.Printf("failed to read registers %d: %v\n", j, err)
			}

			int32Value := int32(regs[0])<<16 | int32(uint16(regs[1]))
			r.Res[i] = int(int32Value)

		default:
			fmt.Printf("failed to read registers data type")
			os.Exit(1)
		}

	}

	r.Name = c.Name
}

func (c *Conf) Decode() {

	file, err := os.Open("conf.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, record := range records {

		//decode le nom
		c.Name = append(c.Name, record[0])

		//decode le registre
		k, err := strconv.Atoi(record[1])
		if err != nil {
			fmt.Printf("failed to convert string to int %v\n", err)
		}
		c.Address = append(c.Address, k)

		//decode la taille de la donnée requise
		c.Size_data = append(c.Size_data, record[2])

		//decode le bit si requis
		j, err := strconv.Atoi(record[3])
		if err != nil {
			fmt.Printf("failed to convert string to int %v\n", err)
		}
		c.Bit = append(c.Bit, j)

		//decode le type de la donnée requise (input, coil ou holding)
		data_type := record[4]

		switch data_type {
		case "coil", "input", "holding":
			c.Type_data = append(c.Type_data, data_type)
		default:
			fmt.Println("wrong input data, must be coil, input or holding")
			os.Exit(1)
		}
	}
}
