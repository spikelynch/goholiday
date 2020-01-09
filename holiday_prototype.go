package main

import (
    "fmt"
    "net"
    "os"
    "bytes"
    "encoding/binary"
    "math/rand"
    "time"
)


type Holiday struct {
	Header [10]uint8
	Globes [150]uint8
}


func rand64() uint8 {
    i := rand.Intn(64)
    return uint8(i)
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


    rand.Seed(time.Now().Unix())

    for {

    	for i := 0; i < 50; i++ {
    		hol.Globes[i * 3] = rand64()
    		hol.Globes[i * 3 + 1] = rand64()
    		hol.Globes[i * 3 + 2] = rand64()
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
    	time.Sleep(10 * time.Millisecond)
    }
    fmt.Println("datagram sent, it's all good")
}
