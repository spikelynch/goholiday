package main

import (
    "fmt"
    "net"
    "os"
    "bytes"
    "encoding/binary"
    "time"
    "math/rand"
    "github.com/lucasb-eyer/go-colorful"

)

/*

TODO -

take the code for the map and holiday lights out to its own module


colour blending

circles grow, shrink, disappear, get replaced by new ones

'feather' edges of circles 





*/

const W2 = 3
const W = 6
const H2 = 4
const H = 8

const VMAX = 0.1

const SLEEP = 50

const RESET = 800


type Holiday struct {
	Header [10]uint8
	Globes [150]uint8
}


type Circle struct {
    x, y, vx, vy, r, r2 float64
    c colorful.Color
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




func rndColor() colorful.Color {
    hue := rand.Float64() * 360
    return colorful.Hsv(hue, 1, 1)
}


func circles(n int) []Circle {
    circles := make([]Circle, n)
    for i := 0; i < n; i++ {
        circles[i].x = rand.Float64() * 6
        circles[i].y = rand.Float64() * 8
        circles[i].vx = 2 * VMAX * rand.Float64() - VMAX 
        circles[i].vy = 2 * VMAX * rand.Float64() - VMAX
        circles[i].r = rand.Float64() * 2 + 0.4
        circles[i].r2 = circles[i].r * circles[i].r
        circles[i].c = rndColor()
    }
    return circles
}


func pointInside(c Circle, x, y float64) bool {
    var dx1, dy1, dx2, dy2 float64
    dx1 = x - c.x
    if x < W2 {
        dx2 = dx1 + W
    } else {
        dx2 = dx1 - W
    }
    dy1 = y - c.y
    if y < H2 {
        dy2 = dy1 + H
    } else {
        dy2 = dy1 - H
    }
    x1 := dx1 * dx1
    x2 := dx2 * dx2
    y1 := dy1 * dy1
    y2 := dy2 * dy2
    return ( x1 + y1 < c.r2 || x1 + y2 < c.r2 || x2 + y1 < c.r2 || x2 + y2 < c.r2 )
}


func calcCircles(bg colorful.Color, circles []Circle, x, y int) colorful.Color {
    for _, c := range circles {
        if pointInside(c, float64(x), float64(y)) {
            return c.c
        }
    }
    return bg
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

    var bg colorful.Color
    var cset []Circle

    for {
        if tick == 0 {
            bg = rndColor()
            cset = circles(rand.Intn(4) + 1)
        }

        for y, row := range m {
            for x, globe := range row {
                cc := calcCircles(bg, cset, x, y)
                setHolidayGlobe(hol, globe, cc)
            }
        }
        sendHoliday(c, hol)

        time.Sleep(SLEEP * time.Millisecond)

        tick += 1
        if tick > RESET {
            tick = 0
        }

        for i, c := range cset {
            c.x = c.x + c.vx
            if c.x < 0 {
                c.x = 5
            }
            if c.x > 5 {
                c.x = 0
            }
            c.y = c.y + c.vy
            if c.y < 0 {
                c.y = 7
            }
            if c.y > 7 {
                c.y = 0
            }
            cset[i] = c
        }


    }


}
