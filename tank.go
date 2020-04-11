package main

import (
    "fmt"
    "net"
    "os"
    "bytes"
    "encoding/binary"
    "math"
)

const SLEEP = 25

const R = 11.0
const H = 12.0

type Holiday struct {
	Header [10]uint8
	Globes [150]uint8
}


type tankMap [50][3]float64

func makeTank() *tankMap {
    tank := new(tankMap)

    r32 := 0.5 * math.Sqrt(3)

    for i := 0; i < 8; i++ {
        tank[i][0] = R
        tank[i][1] = 0
        tank[i][2] = float64(i * H)
    }

    for i := 8; i < 16; i++ {
        tank[i][0] = R * 0.5
        tank[i][1] = R * r32
        tank[i][2] = float64(15 - i) * H
    }

    tank[16][0] = 0
    tank[16][1] = R * r32
    tank[16][2] = -0.5 * H

    for i := 17; i < 25; i++ {
        tank[i][0] = -R * -0.5
        tank[i][1] = R * r32
        tank[i][2] = float64(i - 17) * H
    }

    for i := 25; i < 33; i++ {
        tank[i][0] = -R
        tank[i][1] = 0
        tank[i][2] = float64(32 - i) * H
    }

    tank[33][0] = -0.75 * R
    tank[33][1] = -0.5 * R * r32
    tank[33][2] = -0.5 * H

    for i := 34; i < 42; i++ {
        tank[i][0] = -R * 0.5
        tank[i][1] = -R * r32
        tank[i][2] = float64(i - 34) * H
    }

    for i := 42; i < 50; i++ {
        tank[i][0] = R * 0.5
        tank[i][1] = -R * r32
        tank[i][2] = float64(49 - i) * H
    }

    return tank
}


func sendHoliday(conn *net.UDPConn, hol *Holiday) {
    datagram := new(bytes.Buffer)
    enc_err := binary.Write(datagram, binary.LittleEndian, hol)
    if enc_err != nil {
        fmt.Println("Encoding failed: ", enc_err)
        return
    }

    _, err := conn.Write(datagram.Bytes())

    if err != nil {
        fmt.Println(err)
    }

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

    tank := makeTank()

    for i := 0; i < 50; i++ {
        hol.Globes[i * 3] = uint8(63 * (tank[i][0] + R) / (2 * R))
        hol.Globes[i * 3 + 1] = uint8(63 * (tank[i][1] + R) / (2 * R))
        hol.Globes[i * 3 + 2] = uint8(63 * tank[i][2] / (7 * H))
    }

    sendHoliday(c, hol);

}
