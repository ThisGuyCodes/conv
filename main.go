package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

var (
	from       = flag.String("from", "", "Convert from...")
	fromFormat = flag.String("fromFormat", "", "Format of from file")
	to         = flag.String("to", "", "Convert to...")
	toFormat   = flag.String("toFormat", "", "Desired format of the output")

	nonStreamableFormats = map[string]bool{
		"yaml":        true,
		"json-pretty": true,
		"xml-pretty":  true,
	}
)

type decoder interface {
	Decode(interface{}) error
	More() bool
}

type encoder interface {
	Encode(interface{}) error
}

func createDecoder(format string, reader io.Reader) decoder {
	switch format {
	case "json":
		return json.NewDecoder(reader)
	case "xml":
		return xml.NewDecoder(reader)
	default:
		log.Fatalln("Unknown decoder format:", format)
	}
	return nil
}

func createEncoder(format string, writer io.Writer) encoder {
	switch format {
	case "json":
		return json.NewEncoder(writer)
	case "xml":
		return xml.NewEncoder(writer)
	default:
		log.Fatalln("Unknown encoder format:", format)
	}
	return nil
}

func getMarshaler(format string) func(interface{}) ([]byte, error) {
	switch format {
	case "yaml":
		return yaml.Marshal
	case "json-pretty":
		return func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "  ")
		}
	case "xml-pretty":
		return func(v interface{}) ([]byte, error) {
			return xml.MarshalIndent(v, "", "    ")
		}
	default:
		log.Fatalln("Unknown marshaler format:", format)
	}
	return nil
}

func getUnmarshaler(format string) func([]byte, interface{}) error {
	switch format {
	case "yaml":
		return yaml.Unmarshal
	default:
		log.Fatalln("Unknown unmarshaler format:", format)
	}
	return nil
}

func main() {
	flag.Parse()

	var err error

	var fromFile *os.File
	var toFile *os.File

	if *from == "" {
		log.Fatalln("Please specificy a from file")
	} else if *from == "-" {
		fromFile = os.Stdin
	} else {
		fromFile, err = os.Open(*from)
		if err != nil {
			log.Fatalln("Could not open from file:", err)
		}
	}

	if *to == "" {
		log.Fatalln("Please specify a to file")
	} else if *from == "-" {
		toFile = os.Stdout
	} else {
		toFile, err = os.Create(*to)
		if err != nil {
			log.Fatalln("Could not open to file:", err)
		}
	}

	var fromDatas []map[string]interface{}

	if nonStreamableFormats[*fromFormat] {
		fromUnmarshaler := getUnmarshaler(*fromFormat)
		fromData, err := ioutil.ReadAll(fromFile)
		if err != nil {
			log.Fatalln("Could not read from file:", err)
		}
		var fromType map[string]interface{}
		err = fromUnmarshaler(fromData, &fromType)
		if err != nil {
			log.Fatalln("Could not unmarshal from file:", err)
		}
		fromDatas = append(fromDatas, fromType)
	} else {
		fromDecoder := createDecoder(*fromFormat, fromFile)

		for fromDecoder.More() {
			var fromType map[string]interface{}
			err = fromDecoder.Decode(&fromType)
			if err != nil {
				log.Fatalln("Could not decode from file:", err)
			}
			fromDatas = append(fromDatas, fromType)
		}
	}

	if nonStreamableFormats[*toFormat] {
		toMarshaler := getMarshaler(*toFormat)
		for _, item := range fromDatas {
			toData, err := toMarshaler(item)
			if err != nil {
				log.Fatalln("Could not marshal data:", err)
			}
			for i := 0; i != len(toData) && err == nil; i, err = toFile.Write(toData) {
				toData = toData[i:]
			}
			if err != nil {
				log.Fatalln("Could not write to file:", err)
			}
		}
	} else {
		toEncoder := createEncoder(*toFormat, toFile)

		for _, item := range fromDatas {
			err = toEncoder.Encode(item)
			if err != nil {
				log.Fatalln("Could not encode item:", err)
			}
		}
	}

	err = fromFile.Close()
	if err != nil {
		log.Fatalln("Could not close from file:", err)
	}
	err = toFile.Close()
	if err != nil {
		log.Fatalln("Could not close to file:", err)
	}
}
