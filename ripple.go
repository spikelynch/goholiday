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
    "image/color"
    "image/png"
)

// TODO:
// - put the parameters into a config file
// - make a lightrunner program which can call in different pluggable
//   visualisations as per config - the lightrunner handles the general
//   stuff like the map (lights shape) and when it switches off at night
// - make the map stuff more flexible - you should be able to pass in a 
//   map with config

const NIGHTTIME = 22
const WAKETIME = 6



const R = 11.0
const H = 12.0

const VMAX = 5

const SLEEP = 50

const FADE = 10
const RESET = 2000
const PNGMOD = 1000
const SCREENSHOTS = false

const PI2 = math.Pi * 2

const PNGTILE = 50


type Holiday struct {
	Header [10]uint8
	Globes [150]uint8
}



// func makeMap_orig() [][]int {
//     m := make([][]int, 8)

//     for i := 0; i < 8; i++ {
//         m[i] = make([]int, 6)
//         m[i][0] = i;
//         m[i][1] = 15 - i;
//         m[i][2] = 17 + i;
//         m[i][3] = 32 - i;
//         m[i][4] = 34 + i;
//         m[i][5] = 49 - i;
//     }

//     return m

// }

func makeMap() [][]int {
    m := make([][]int, 5)

    for i := 0; i < 5; i++ {
        m[i] = make([]int, 10)
        m[i][0] = i;
        m[i][1] = 9 - i;
        m[i][2] = 10 + i;
        m[i][3] = 19 - i;
        m[i][4] = 20 + i;
        m[i][5] = 29 - i;
        m[i][6] = 30 + i;
        m[i][7] = 39 - i;
        m[i][8] = 40 + i;
        m[i][9] = 49 - i;
        fmt.Printf("%d: %v\n", i, m[i]);
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


func setHolidayDark(conn *net.UDPConn, hol *Holiday) {
    for i := 0; i < 50; i++ {
        setHolidayGlobe(hol, i, colorful.Color{0, 0, 0})
    }
    sendHoliday(conn, hol)
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


func hueRange(hue, spread, dim, sat, value, z float64) colorful.Color {
    h := hue + spread * z
    if h < 0 {
        h = h + 360
    } else if h > 360 {
    h = h - 360
    }
    return colorful.Hsv(h, sat, value * (dim * 0.5 * (z + 1) + 1 - dim))
}



func sinusoidal(x, y, xfreq, yfreq, xphase, yphase, twist, curl float64) float64 {
    theta := xphase + x * xfreq / 6.0
    phi := yphase + (y - 4) * yfreq / 7.0
    pt := PI2 * (theta + twist * math.Sin(curl * phi))
    pp := PI2 * phi
    s := math.Sin(pt) * math.Sin(pp)
    return s
}



func screenshot(run, frame int, xfreq, yfreq, xphase, yphase, twist, curl, hue, spread, dim, sat float64) {
    img := image.NewNRGBA(image.Rect(0, 0, PNGTILE * 6, PNGTILE * 7))

    for y := 0; y < PNGTILE * 7; y++ {
        for x := 0; x < PNGTILE * 6; x++ {
            z := sinusoidal(float64(x) / float64(PNGTILE), float64(y) / float64(PNGTILE), xfreq, yfreq, xphase, yphase, twist, curl)
            pixel := hueRange(hue, spread, dim, sat, 1.0, z)
            img.Set(x, y, color.NRGBA{
                R: uint8(pixel.R * 255),
                G: uint8(pixel.G * 255),
                B: uint8(pixel.B * 255),
                A: 255,
            })
        }
    }
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
    run := 0
    frame := 0

    var hue, hdrift, spread, sat, dim float64
    var fade float64
    var xfreq, yfreq, yfreqamp, yfreqmean, yfreqvel float64
    var xvel, yvel, twist, curl float64

    for {

        if tick == 0 {

            tn := time.Now()
            hour := tn.Hour()

            if hour == NIGHTTIME {
                setHolidayDark(c, hol)
                sleeptime := (24 - NIGHTTIME + WAKETIME)
                fmt.Printf("sleeping for %d hours\n", sleeptime)
                time.Sleep(time.Duration(sleeptime) * time.Hour)
            }

            hue = 360.0 * rand.Float64()
	    hdrift = rand.Float64() * 1 - 0.5
            spread = 180.0 //40.0 + 140.0 * rand.Float64()
	    sat = 0.6 + 0.4 * rand.Float64()
	    dim = rand.Float64()

            xfreq = float64(rand.Intn(2) + 1)
            yfreqmean = float64(rand.Intn(2) + 1)
            yfreqamp = rand.Float64() * yfreqmean
            yfreqvel = rand.Float64() * 0.05

            twist = rand.Float64() * 2.0 - 1.0
            curl = rand.Float64() * 5 - 2.5

            xvel = rand.Float64() * 0.02 - 0.04
            yvel = rand.Float64() * 0.02 - 0.04

        }

        xphase := xvel * float64(tick)
        yphase := yvel * float64(tick)

        yfreq = yfreqmean + yfreqamp * math.Sin(float64(tick) * yfreqvel)

        if tick < FADE {
            fade = float64(tick) / float64(FADE)
        } else if tick > RESET - FADE {
            fade = float64(RESET - tick) / float64(FADE)
        } else {
            fade = 1.0
        }

        for y, row := range m {
            for x, globe := range row {
                z := sinusoidal(float64(x), float64(y), xfreq, yfreq, xphase, yphase, twist, curl)
                colour := hueRange(hue, spread, dim, sat, fade, z)
                setHolidayGlobe(hol, globe, colour)
            }
        }

        hue += hdrift
	if hue < 0 {
	    hue += 360
	} else if hue > 360 {
	    hue -= 360
	}
        sendHoliday(c, hol)
        time.Sleep(SLEEP * time.Millisecond)
        if SCREENSHOTS {
            if tick == 0 {
                screenshot(run, frame, xfreq, yfreq, xphase, yphase, twist, curl, hue, spread, dim, sat)
                frame += 1
            }
        }

        tick += 1
        if tick > RESET {
            tick = 0
            frame = 0
            run += 1
        }

    }

}
