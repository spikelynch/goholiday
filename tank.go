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

const SLEEP = 25

const SPEED = 0.05

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


type tankMap [50]Point3D


func makeTank() *tankMap {
    tank := new(tankMap)

    r32 := 0.5 * math.Sqrt(3)

    for i := 0; i < 8; i++ {
        tank[i] = Point3D{ R, 0, float64(i * H) }
    }

    for i := 8; i < 16; i++ {
        tank[i] = Point3D{ R * 0.5, R * r32, float64(15 - i) * H }
    }

    tank[16] = Point3D{ 0, R * r32, -0.5 * H }

    for i := 17; i < 25; i++ {
        tank[i] = Point3D{ -R * -0.5, R * r32, float64(i - 17) * H }
    }

    for i := 25; i < 33; i++ {
        tank[i] =  Point3D{ -R, 0, float64(32 - i) * H }
    }

    tank[33] = Point3D{ -0.75 * R, -0.5 * R * r32, -0.5 * H }

    for i := 34; i < 42; i++ {
        tank[i] = Point3D{ -R * 0.5, -R * r32, float64(i - 34) * H }
    }

    for i := 42; i < 50; i++ {
        tank[i] = Point3D{ R * 0.5, -R * r32, float64(49 - i) * H }
    }

    return tank
}


func showHoliday(hol *Holiday) {
    for i := 0; i < 50; i++ {
        fmt.Printf("hol %d = %d %d %d\n", i, hol.Globes[i * 3], hol.Globes[i * 3 + 1], hol.Globes[i * 3 + 2])
    }
}

func setHolidayGlobe(hol *Holiday, i int, c Colour) {
    hol.Globes[i * 3] = uint8(63 * c.r)
    hol.Globes[i * 3 + 1] = uint8(63 * c.g)
    hol.Globes[i * 3 + 2] = uint8(63 * c.b)
}



func illuminate(light *Light, point Point3D) Colour {
    dx := (light.p.x - point.x)
    dy := (light.p.y - point.y)
    dz := (light.p.z - point.z)
    d2 := dx * dx + dy * dy + dz * dz
    lumen := 1.0 / (OFFSET + d2 * DIMMING)
    return Colour{ light.c.r * lumen, light.c.g * lumen, light.c.b * lumen }
}


func illuminateTank(lights []Light, tank *tankMap) [50]Colour {
    var colours [50]Colour
    n := float64(len(lights))
    for i := 0; i < 50; i++ {
        c := Colour{0,0,0}
        for _, light := range lights {
            cl := illuminate(&light, tank[i])
            c.r += cl.r
            c.g += cl.g
            c.b += cl.b
        }
        colours[i] = Colour{c.r / n, c.g / n, c.b / n}
    }
    return colours
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


func randomVelocity() Point3D {
    theta := rand.Float64() *  2 * math.Pi
    phi := ( 0.4 + 0.1 * rand.Float64()) * math.Pi
    v := (0.6 + 0.4 * rand.Float64()) * VMAX
    vx := v * math.Cos(phi) * math.Cos(theta)   
    vy := v * math.Cos(phi) * math.Sin(theta)   
    vz := v * math.Sin(phi)
    return Point3D{ vx, vy, vz }
}


func randomColour() Colour {
    c := colorful.Hsv(rand.Float64() * 360.0, 1, 1)
    fmt.Printf("RGB %f %f %f\n", c.R, c.G, c.B)
    return Colour{ c.R, c.G, c.B }
}

func outOfBounds(p Point3D) bool {
    return ( p.x < -2 * R || p.x > 2 * R || p.y < -2 * R || p.y > 2 * R || p.z > 9 * H)
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

    tank := makeTank()

    lights := make([]Light, NLIGHTS)
    lightvs := make([]Point3D, NLIGHTS)

    for i := 0; i < NLIGHTS; i++ {
        lights[i] = Light{ Point3D{ 0, 0, - 0.5 * H }, randomColour() }
        lightvs[i] = randomVelocity()
    }


    for {
        for i := 0; i < NLIGHTS; i++ {
            lights[i].p.x += lightvs[i].x
            lights[i].p.y += lightvs[i].y
            lights[i].p.z += lightvs[i].z
            if outOfBounds(lights[i].p) {
                lights[i] = Light{ Point3D{ 0, 0, -0.5 * H }, randomColour() }
                lightvs[i] = randomVelocity()
            }
        }
        globes := illuminateTank(lights, tank)
        for i := 0; i < 50; i++ {
            setHolidayGlobe(hol, i, globes[i])
        }
        sendHoliday(c, hol)        
        time.Sleep(SLEEP * time.Millisecond)
    }


}
