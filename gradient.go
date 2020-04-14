package main

import (
    "fmt"
    "net"
    "os"
    "bytes"
    "encoding/binary"
    "math/rand"
    "time"
    "github.com/lucasb-eyer/go-colorful"
)

const SLEEP = 25

const SPEEDTOP = 0.23
const SPEEDBOT = -0.31

const R = 11.0
const H = 12.0

const NLIGHTS = 25
const VMAX = 5

const OFFSET = 0.5
const DIMMING = 0.03

type Holiday struct {
	Header [10]uint8
	Globes [150]uint8
}

type Point3D struct {
    x, y, z float64
}

type Colour struct {
    r, g, b float64
}

type Light struct {
    p Point3D
    c Colour
}





func makeMap() [][]int {
    m := make([][]int, 9)
    m[0] = make([]int, 2)
    m[0][0] = 16
    m[0][1] = 33

    for i := 1; i < 9; i++ {
        m[i] = make([]int, 6)
        m[i][0] = -1 + i;
        m[i][1] = 16 - i;
        m[i][2] = 16 + i;
        m[i][3] = 33 - i;
        m[i][4] = 33 + i;
        m[i][5] = 50 - i;
    }

    return m

}


func showHoliday(hol *Holiday) {
    for i := 0; i < 50; i++ {
        fmt.Printf("hol %d = %d %d %d\n", i, hol.Globes[i * 3], hol.Globes[i * 3 + 1], hol.Globes[i * 3 + 2])
    }
}

func setHolidayGlobe(hol *Holiday, i int, c colorful.Color) {
    hol.Globes[i * 3] = uint8(63 * c.R)
    hol.Globes[i * 3 + 1] = uint8(63 * c.G)
    hol.Globes[i * 3 + 2] = uint8(63 * c.B)
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




func randomColour() Colour {
    c := colorful.Hsv(rand.Float64() * 360.0, 1, 1)
    fmt.Printf("RGB %f %f %f\n", c.R, c.G, c.B)
    return Colour{ c.R, c.G, c.B }
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

    fmt.Printf("Holiday is %s\n", c.RemoteAddr().String())
    defer c.Close()

    rand.Seed(time.Now().UnixNano())

    hol := new(Holiday)

    m := makeMap()

    htop := 0.0
    hbot := 180.0

    for {

        cbot := colorful.Hsv(hbot, 1, 1)
        ctop := colorful.Hsv(htop, 1, 1)

        for i, row := range m {
            k := float64(i)/float64(len(m) - 1)
            gradc := cbot.BlendHcl(ctop, k).Clamped()
            for _, globe := range row {
                setHolidayGlobe(hol, globe, gradc)
            }
        }
        sendHoliday(c, hol)
        time.Sleep(SLEEP * time.Millisecond)

        hbot += SPEEDBOT
        if hbot > 360.0 {
            hbot -= 360.0
        } else if hbot < 0.0 {
            hbot += 360.0
        }
        htop += SPEEDTOP
        if htop > 360.0 {
            htop -= 360.0
        } else if htop < 0.0 {
            htop += 360.0
        }
    }
}
