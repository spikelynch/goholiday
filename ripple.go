package main

import (
    "fmt"
    "net"
    "os"
    "bytes"
    "encoding/binary"
    "time"
    "math"
    "math/rand"
    "github.com/lucasb-eyer/go-colorful"
)


const R = 11.0
const H = 12.0

const NLIGHTS = 25
const VMAX = 5

const SLEEP = 50

const FADE = 100

const RESET = 2000

const PI2 = math.Pi * 2


type Holiday struct {
	Header [10]uint8
	Globes [150]uint8
}



func makeMap() [][]int {
    m := make([][]int, 8)

    for i := 0; i < 8; i++ {
        m[i] = make([]int, 6)
        m[i][0] = i;
        m[i][1] = 15 - i;
        m[i][2] = 17 + i;
        m[i][3] = 32 - i;
        m[i][4] = 34 + i;
        m[i][5] = 49 - i;
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




func colorPair() []colorful.Color {
    hueOn := rand.Float64() * 360
    hueOff := hueOn + 120 + rand.Float64() * 120
    if hueOff > 360 {
        hueOff -= 360
    } else if hueOff < 0 {
        hueOff += 360
    }
    pair := make([]colorful.Color, 2)
    pair[0] = colorful.Hsv(hueOff, 1, 0.5 * rand.Float64())
    pair[1] = colorful.Hsv(hueOn, 1, 0.5 + 0.5 * rand.Float64())
    return pair
}



func toColour(cols []colorful.Color, z float64) colorful.Color {
    return cols[0].BlendHcl(cols[1], 0.5 * (z + 1)).Clamped()
}

func hueRange(hue, spread, value, z float64) colorful.Color {
    h := hue + spread * z
    if h < 0 {
        h = h + 360
    } else if h > 360 {
	h = h - 360
    }
    return colorful.Hsv(h, 1, value)
}



func sinusoidal(x, y, xfreq, yfreq int, xphase, yphase, twist float64) float64 {
    theta := (xphase + float64(x * xfreq) / 6.0)
    phi := (yphase + float64(y * yfreq) / 7.0)
    pt := PI2 * (theta + twist * phi)
    pp := PI2 * phi
    s := math.Sin(pt) * math.Sin(pp)
    return s
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


    tick := 0

    //var cols []colorful.Color
    var xfreq, yfreq int
    var xvel, yvel, twist float64
    var hue, spread, cvel float64
    var fade float64

    for {
        if tick == 0 {
            //cols = colorPair()

	    hue = 360.0 * rand.Float64()
	    spread = 40.0 + 50.0 * rand.Float64()
	    cvel = 0 //rand.Float64() * 2.0 - 1.0
            xfreq = 1 + rand.Intn(3)
            yfreq = 1 + rand.Intn(6)
            twist = rand.Float64() * 8.0 - 4.0

            xvel = rand.Float64() * 0.04 - 0.02
            yvel = rand.Float64() * 0.04 - 0.02

        }

	if tick < FADE {
	    fade = float64(tick) / float64(FADE)
	} else if tick > RESET - FADE {
	    fade = float64(RESET - tick) / float64(FADE)
	} else {
            fade = 1.0
	}

        xphase := xvel * float64(tick)
        yphase := yvel * float64(tick)

	hue = hue + cvel
	if hue > 360 {
	    hue = hue - 360
	} else if hue < 0 {
            hue = hue + 360
	}

        for y, row := range m {
            for x, globe := range row {
		z := sinusoidal(x, y, xfreq, yfreq, xphase, yphase, twist)
		colour := hueRange(hue, spread, fade, z)
                setHolidayGlobe(hol, globe, colour)
            }
        }
        sendHoliday(c, hol)
        time.Sleep(SLEEP * time.Millisecond)

        tick += 1
        if tick > RESET {
            tick = 0
        }

    }   

}
