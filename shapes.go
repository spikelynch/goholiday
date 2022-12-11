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

/*

TODO -

take the code for the map and holiday lights out to its own module

circles grow, shrink, disappear, get replaced by new ones

*/

const W2 = 5
const W = 10
const H2 = 2.5
const H = 5

const VXMIN = -0.001
const VXMAX = 0.001

const VYMIN = 0.1
const VYMAX = 0.12

const VYMIN = 0.01
const VYMAX = 0.04

const NMIN = 2
const NMAX = 5

const RMIN = 0.5
const RMAX = 1.0


const LMIN = 2400
const LMAX = 4800


const SLEEP = 25

const BMIN = 400
const BMAX = 700

const BG_SAT = 0.5
const BG_VALUE = 0.4

const FUZZ = 1.5


type Holiday struct {
	Header [10]uint8
	Globes [150]uint8
}


type Circle struct {
    x, y, vx, vy, r, hue float64
    lifespan, t, pulse int
}




// func makeMap() [][]int {
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


func randInt(min, max int) int {
    if max <= min {
        return min
    } else {
        return min + rand.Intn(max - min)
    }
}


func rndCircle() Circle {
    c := new(Circle)
    c.x = rand.Float64() * 6
    c.y = rand.Float64() * 8
    c.vx = VXMIN + rand.Float64() * (VXMAX - VXMIN)
    c.vy = VYMIN + rand.Float64() * (VYMAX - VYMIN)
    c.r = RMIN + rand.Float64() * (RMAX - RMIN)
    c.hue = rand.Float64() * 360
    c.lifespan = randInt(LMIN, LMAX)
    c.t = 0
    c.pulse = 1
    return *c
}



func circles(n int) []Circle {
    circles := make([]Circle, n)
    for i := 0; i < n; i++ {
        circles[i] = rndCircle()
    }
    return circles
}


// func lifecycle(c Circle) float64 {
//     if c.t < c.lifespan / 2 {
//         return c.r * 2 * float64(c.t) / float64(c.lifespan)
//     } else {
//         return c.r * 2 * float64(c.lifespan - c.t) / float64(c.lifespan)
//     }
// }


func lifecycle(c Circle) float64 {
    return c.r * math.Sin(math.Pi * float64(c.t) / float64(c.lifespan))
}


func fuzz(d2, r1, r2 float64) float64 {
    if d2 < r1 {
        return 1
    }
    if d2 < r2 {
        return 1 - (d2 - r1) / (r2 - r1) 
    }
    return 0
}




func circleValue(c Circle, x, y float64) float64 {
    var dx1, dy1, dx2, dy2 float64
    rnow := lifecycle(c)
    r2 := rnow * rnow
    fr1 := rnow - FUZZ
    r1 := fr1 * fr1
    if fr1 < 0 {
        r1 = -r1
    } 
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
    v := fuzz(x1 + y1, r1, r2) + fuzz(x1 + y2, r1, r2) + fuzz(x2 + y1, r1, r2) + fuzz(x2 + y2, r1, r2)
    if v < 1 {
        return v
    }
    return 1
}


func renderCircles(bg colorful.Color, circles []Circle, x, y int) colorful.Color {
    r := bg.R
    g := bg.G
    b := bg.B
    for _, c := range circles {
        v := circleValue(c, float64(x), float64(y))
        cc := colorful.Hsv(c.hue, 1, v)
        r += cc.R
        g += cc.G
        b += cc.B
    }
    if r > 1 {
        r = 1
    }
    if g > 1 {
        g = 1
    }
    if b > 1 {
        b = 1
    }
    return colorful.Color{r, g, b}
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
	fmt.Println("Could not connect")
        fmt.Println(err)
        return
    }

    fmt.Printf("Holiday is %s\n", c.RemoteAddr().String())
    defer c.Close()

    rand.Seed(time.Now().UnixNano())

    hol := new(Holiday)

    m := makeMap()


    tick := 0

    nextbirth := randInt(BMIN, BMAX)

    bg := colorful.Hsv(rand.Float64() * 360, BG_SAT, BG_VALUE)
    cset := []Circle{}

    for {
        if tick == 0 {
            cset = append(cset, rndCircle())
            //fmt.Println("a circle was born")
        }

        for y, row := range m {
            for x, globe := range row {
                cc := renderCircles(bg, cset, x, y)
                setHolidayGlobe(hol, globe, cc)
            }
        }
        sendHoliday(c, hol)

        time.Sleep(SLEEP * time.Millisecond)

        tick += 1
        if tick > nextbirth {
            tick = 0
            nextbirth = randInt(BMIN, BMAX)
        }

        csetnext := []Circle{}

        for _, c := range cset {
            c.x = c.x + c.vx
            if c.x < 0 {
                c.x = c.x + W
            }
            if c.x > W {
                c.x = c.x - W
            }
            c.y = c.y + c.vy
            if c.y < 0 {
                c.y = c.y + H
            }
            if c.y > H {
                c.y = c.y - H
            }
            c.t += 1
            if c.t < c.lifespan {
                csetnext = append(csetnext, c)   
            } 
        }

        cset = csetnext

    }


}
