package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

var (
	inputPath  string
	outputPath string
)

func init() {
	flag.StringVar(&inputPath, "in", "-", "Path to the input file")
	flag.StringVar(&outputPath, "out", "-", "Path to the output file")
}

type holdType struct {
	held interface{}
}

func (ht *holdType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var held interface{}
	err := unmarshal(&held)
	if err != nil {
		return err
	}

	switch held := held.(type) {
	case map[interface{}]interface{}:
		strHeld := make(map[string]interface{}, len(held))
		for k, v := range held {
			switch k := k.(type) {
			case string:
				strHeld[k] = v
			case int:
				strHeld[strconv.Itoa(k)] = v
			}
		}
		ht.held = strHeld
		return nil
	default:
		ht.held = held
		return nil
	}
}

func (ht *holdType) MarshalYAML() (interface{}, error) {
	return ht.held, nil
}

func (ht *holdType) MarshalJSON() ([]byte, error) {
	return json.Marshal(ht.held)
}

func (ht *holdType) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &ht.held)
}

func main() {
	var err error
	flag.Parse()

	var inputFile *os.File
	if inputPath == "-" {
		inputFile = os.Stdin
	} else {
		inputFile, err = os.Open(inputPath)
		if err != nil {
			log.Fatalln("Could not open input file:", err)
		}
	}
	defer inputFile.Close()

	var outputFile *os.File
	if outputPath == "-" {
		outputFile = os.Stdout
	} else {
		outputFile, err = os.Open(outputPath)
		if err != nil {
			log.Fatalln("Could not open input file:", err)
		}
	}
	defer outputFile.Close()

	inputData, err := ioutil.ReadAll(inputFile)
	if err != nil {
		log.Fatalln("Could not read input file:", err)
	}

	inputInterface := &holdType{}
	err = yaml.Unmarshal(inputData, &inputInterface)
	if err != nil {
		log.Fatalln("Could not unmarshal input data:", err)
	}

	outputData, err := json.MarshalIndent(inputInterface, "", "  ")
	if err != nil {
		log.Fatalln("Could not marshal the output:", err)
	}

	i, err := outputFile.Write(outputData)
	if err != nil {
		log.Fatalln("Could not write to output file:", err)
	}
	for i != len(outputData) {
		outputData = outputData[i:]
		i, err = outputFile.Write(outputData)
		if err != nil {
			log.Fatalln("Could not write to output file:", err)
		}
	}
}
