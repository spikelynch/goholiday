package main

import (
    "fmt"
    "net"
    "os"
    "bytes"
    "encoding/binary"
    "math"
    "time"
)

const SLEEP = 25
const REDPHASE = 0.01
const GREENPHASE = 0.013
const BLUEPHASE = 0.009

type Holiday struct {
	Header [10]uint8
	Globes [150]uint8
}


func phaser(phase float64, i int) uint8 {
    return uint8(32 + 31 * math.Cos(float64(i) * math.Sin(phase)))
}



func main() {
    arguments := os.Args
    if len(arguments) == 1 {
        fmt.Println("Please provide a host:port string")
        return
    }
    CONNECT := arguments[1]

    s, err := net.ResolveUDPAddr("udp4", CONNECT)
    c, err := net.DialUDP("udp4", nil, s)
    if err != nil {
        fmt.Println(err)
        return
    }

    fmt.Printf("The UDP server is %s\n", c.RemoteAddr().String())
    defer c.Close()

    hol := new(Holiday)

    phase := 0

    for {
        phase += 1
    	for i := 0; i < 50; i++ {
    		hol.Globes[i * 3] = phaser(REDPHASE * float64(phase), i)
    		hol.Globes[i * 3 + 1] = phaser(GREENPHASE * float64(phase), i)
    		hol.Globes[i * 3 + 2] = phaser(BLUEPHASE * float64(phase), i)
    	}

    	datagram := new(bytes.Buffer)
    	enc_err := binary.Write(datagram, binary.LittleEndian, hol)
    	if enc_err != nil {
    		fmt.Println("Encoding failed: ", enc_err)
    		return
    	}


    	_, err = c.Write(datagram.Bytes())

    	if err != nil {
        	fmt.Println(err)
        	
    	}
    	time.Sleep(SLEEP * time.Millisecond)
    }
    fmt.Println("datagram sent, it's all good")
}
