package main

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/gopacket/pcapgo"
)

func packetCount(pcapPath string) (uint64, error) {
	inputFile, err := os.Open(pcapPath)
	if err != nil {
		return 0, err
	}
	defer inputFile.Close()

	reader, err := pcapgo.NewReader(inputFile)
	if err != nil {
		return 0, err
	}

	var packetCount uint64
	for {
		_, _, err = reader.ReadPacketData()
		if err == io.EOF {
			return packetCount, nil
		} else if err != nil {
			return 0, err
		}
		packetCount++
	}
}

func isTimeOrdered(pcapPath string) (bool, error) {
	inputFile, err := os.Open(pcapPath)
	if err != nil {
		return false, err
	}
	defer inputFile.Close()

	reader, err := pcapgo.NewReader(inputFile)
	if err != nil {
		return false, err
	}

	var previousTime int64
	for {
		_, capInfo, err := reader.ReadPacketData()
		if err == io.EOF {
			return true, nil
		} else if err != nil {
			return false, err
		}

		currentTime := capInfo.Timestamp.UnixNano()

		if currentTime < previousTime {
			return false, nil
		}

		previousTime = currentTime
	}
}

// TestCount packet count of merged pcap
// should be the sum of the packet counts of the
// input pcaps
func TestCount(t *testing.T) {
	outputFile, err := ioutil.TempFile("", "joincap_output_")
	if err != nil {
		t.Fatal(err)
	}
	outputFile.Close()
	defer os.Remove(outputFile.Name())

	inputFilePath := "pcap_examples/ok.pcap"

	joincap([]string{"joincap", "-w", outputFile.Name(), inputFilePath, inputFilePath})

	inputPacketCount, err := packetCount(inputFilePath)
	if err != nil {
		t.Fatal(err)
	}
	outputPacketCount, err := packetCount(outputFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if inputPacketCount*2 != outputPacketCount {
		t.Fatalf("inputPacketCount*2 != outputPacketCount (%d != %d)\n", inputPacketCount*2, outputPacketCount)
	}
}

// TestOrder all packets in merged pacap should
// be ordered by time
func TestOrder(t *testing.T) {
	outputFile, err := ioutil.TempFile("", "joincap_output_")
	if err != nil {
		t.Fatal(err)
	}
	outputFile.Close()
	defer os.Remove(outputFile.Name())

	inputFilePath := "pcap_examples/ok.pcap"

	joincap([]string{"joincap", "-w", outputFile.Name(), inputFilePath, inputFilePath})

	isInputOrdered, err := isTimeOrdered(inputFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if !isInputOrdered {
		t.Fatalf("inputFile %s is not ordered by time\n", inputFilePath)
	}

	isOutputOrdered, err := isTimeOrdered(outputFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !isOutputOrdered {
		t.FailNow()
	}
}