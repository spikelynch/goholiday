package main

import (
    "fmt"
    "net"
    "os"
    "bytes"
    "encoding/binary"
    "math"
    "math/rand"
    "time"
    "github.com/lucasb-eyer/go-colorful"
)


const R = 11.0
const H = 12.0

const NLIGHTS = 25
const VMAX = 5

const SLEEP = 25


const BOTRATE = 0.05
const BOTOFF = 0.1
const BOTAMP = 0.1
const BOTSTART = 0.0

const TOPRATE = 0.062
const TOPOFF = 0.1
const TOPAMP = 0.1
const TOPSTART = 120.0


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


func makeGradient(bot, top float64, n int) []colorful.Color {
    gradient := make([]colorful.Color, n)
    cbot := colorful.Hcl(bot, 1, 0.5)
    ctop := colorful.Hcl(top, 1, 0.5)

    for i := 0; i < n; i++ {
        k := float64(i)/float64(n - 1)
        // gradient[i] = cbot.BlendRgb(ctop, k)
        gradient[i] = cbot.BlendHcl(ctop, k).Clamped()
    }

    return gradient
}

func velocity(t int, rate, offset, amp float64) float64 {
    return math.Sin(float64(t) * rate) * amp + offset
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

    htop := TOPSTART 
    hbot := BOTSTART

    tick := 0

    for {

        grad := makeGradient(hbot, htop, len(m))

        for i, row := range m {
            for _, globe := range row {
                setHolidayGlobe(hol, globe, grad[i])
            }
        }
        sendHoliday(c, hol)
        time.Sleep(SLEEP * time.Millisecond)

        hbot += velocity(tick, BOTRATE, BOTOFF, BOTAMP)
        if hbot > 360.0 {
            hbot -= 360.0
        } else if hbot < 0.0 {
            hbot += 360.0
        }
        htop += velocity(tick, TOPRATE, TOPOFF, TOPAMP)
        if htop > 360.0 {
            htop -= 360.0
        } else if htop < 0.0 {
            htop += 360.0
        }
        tick += 1
    }
}
