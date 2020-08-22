package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	thetaSpacing = 0.03
	phiSpacing   = 0.01
	r1           = 1 // radius of the revolved circle (i.e. thickness of torus)
	r2           = 2 // radius of inner torus circle
	k2           = 5 // distance of torus to plance of projection
)

var (
	luminenceChars = []rune{
		'.',
		',',
		'-',
		'~',
		':',
		';',
		'=',
		'!',
		'*',
		'#',
		'$',
		'@',
	}
)

func main() {
	a := 0.0
	b := 0.0

	sigs := make(chan os.Signal, 1)
	caughtSig := false

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		caughtSig = true
	}()

	for {
		render(a, b)
		if caughtSig {
			break
		}
		a += .04
		b += .02
		time.Sleep(10 * time.Millisecond)
	}
}

func render(a, b float64) {
	width, height := screenSize()

	output := make([][]rune, width)
	for i := range output {
		output[i] = make([]rune, height)
		for j := range output[i] {
			output[i][j] = ' ' // initialize with spaces instead of NULL
		}
	}

	zbuffer := make([][]float64, width)
	for i := range zbuffer {
		zbuffer[i] = make([]float64, height)
	}

	cosa, sina := math.Cos(a), math.Sin(a)
	cosb, sinb := math.Cos(b), math.Sin(b)

	k1 := width * k2 * 3 / (8 * (r1 + r2)) // distance of viewer to plane of projection

	for theta := 0.0; theta < 2*math.Pi; theta += thetaSpacing {
		costheta, sintheta := math.Cos(theta), math.Sin(theta)

		// x/y coordinate of the current point on the circle, before revolution and rotation
		circlex := r2 + r1*costheta
		circley := r1 * sintheta

		for phi := 0.0; phi < 2*math.Pi; phi += phiSpacing {
			cosphi, sinphi := math.Cos(phi), math.Sin(phi)

			// now, revolve according to phi and a/b
			x := circlex*(cosb*cosphi+sina*sinb*sinphi) - circley*cosa*sinb
			y := circlex*(sinb*cosphi-sina*cosb*sinphi) + circley*cosa*cosb
			z := k2 + cosa*circlex*sinphi + circley*sina
			zinverse := 1.0 / z

			// project to x/y on the plane
			xproj := (width/2 + int(float64(k1)*zinverse*x))
			yproj := (height/2 - int(float64(k1)*zinverse*y))

			luminence := cosphi*costheta*sinb - cosa*costheta*sinphi - sina*sintheta + cosb*(cosa*sintheta-costheta*sina*sinphi)

			if luminence > 0.0 {
				// zinverse is how close the point is to the viewer.
				// so, if the point is closer than something already
				if zinverse > zbuffer[xproj][yproj] {
					zbuffer[xproj][yproj] = zinverse
					luminenceIdx := int(luminence * 8.0) // ranges from 0..11

					output[xproj][yproj] = luminenceChars[luminenceIdx]
				}
			}
		}
	}

	printOutput(output)
}

func printOutput(output [][]rune) {
	fmt.Print("\x1b[H")
	for i := 0; i < len(output); i++ {
		fmt.Println(string(output[i]))
	}
}

func screenSize() (int, int) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	sizeArr := strings.Split(strings.TrimSpace(string(out)), " ")

	return mustParse(sizeArr[0]), mustParse(sizeArr[1])
}

func mustParse(s string) int {
	x, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		panic(err)
	}
	return int(x)
}
