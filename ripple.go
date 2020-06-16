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
    "image"
 //   "image/color"
    "image/png"
)


const R = 11.0
const H = 12.0

const NLIGHTS = 25
const VMAX = 5

const SLEEP = 50

const RESET = 1000
const PNGMOD = 50

const PI2 = math.Pi * 2

const PNGTILE = 100

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
    pair[0] = colorful.Hsv(hueOff, 1, 1)// 
    pair[1] = colorful.Hsv(hueOn, 1, 1)
    return pair
}



func toColour(cols []colorful.Color, z float64) colorful.Color {
    return cols[0].BlendHcl(cols[1], 0.5 * (z + 1)).Clamped()
}


func sinusoidal_old(x, y, xfreq, yfreq int, xphase, yphase, twist float64) float64 {
    theta := (xphase + float64(x * xfreq) / 6.0)
    phi := (yphase + float64(y * yfreq) / 7.0)
    pt := PI2 * (theta + twist * phi)
    pp := PI2 * phi
    s := math.Sin(pt) * math.Sin(pp)
    return s
}

func sinusoidal(x, y int, xfreq, yfreq, xphase, yphase, twist float64) float64 {
    theta := xphase + float64(x) * xfreq / 6.0
    phi := yphase + float64(y) * yfreq / 7.0
    pt := PI2 * (theta + twist * phi)
    pp := PI2 * phi
    s := math.Sin(pt) * math.Sin(pp)
    return s
}



func screenshot(run, frame int, cols []colorful.Color, xfreq, yfreq int, xphase, yphase, twist float64) {
    img := image.NewNRGBA(image.Rect(0, 0, PNGTILE * 6, PNGTILE * 7))

    // for y := 0; y < PNGTILE * 7; y++ {
    //     for x := 0; x < PNGTILE * 6; x++ {
    //         pixel := toColour(cols, sinusoidal(float64(x) / float64(PNGTILE), float64(y) / float64(PNGTILE), float64(xfreq), float64(yfreq), xphase, yphase, twist))
    //         img.Set(x, y, color.NRGBA{
    //             R: uint8(pixel.R * 255),
    //             G: uint8(pixel.G * 255),
    //             B: uint8(pixel.B * 255),
    //             A: 255,
    //         })
    //     }
    // }
    filename := fmt.Sprintf("./images/r%df%04d.png", run, frame)
    f, err := os.Create(filename)
    if err != nil {
        fmt.Println(err)
    }

    if err := png.Encode(f, img); err != nil {
        f.Close()
        fmt.Println(err)
    }

    if err := f.Close(); err != nil {
        fmt.Println(err)
    }
    fmt.Printf("Wrote screenshot %s\n", filename)
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
    //run := 0
    //frame := 0

    var cols []colorful.Color
    var xfreq, yfreq, yfreqamp, yfreqmean, yfreqvel float64
    var xvel, yvel, twist float64


    for {
        if tick == 0 {
            cols = colorPair()

            xfreq = float64(rand.Intn(4) + 1)
            yfreqmean = float64(rand.Intn(4) + 1)
            yfreqamp = rand.Float64() * yfreqmean
            yfreqvel = rand.Float64() * 0.2

            twist = rand.Float64() * 4.0 - 2.0

            xvel = rand.Float64() * 0.02 - 0.04
            yvel = rand.Float64() * 0.02 - 0.04

        }

        xphase := xvel * float64(tick)
        yphase := yvel * float64(tick)

        yfreq = yfreqmean + yfreqamp * math.Sin(float64(tick) * yfreqvel)

        for y, row := range m {
            for x, globe := range row {
                setHolidayGlobe(hol, globe, toColour(cols, sinusoidal(x, y, xfreq, yfreq, xphase, yphase, twist)))
            }
        }


        sendHoliday(c, hol)
        time.Sleep(SLEEP * time.Millisecond)
        //if tick % PNGMOD == 0 {
        //    screenshot(run, frame, cols, xfreq, yfreq, xphase, yphase, twist)
        //    frame += 1
        //}

        tick += 1
        if tick > RESET {
            tick = 0
            //frame = 0
            //run += 1
        }

    }   

}