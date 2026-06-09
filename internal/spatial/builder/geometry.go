package builder

import (
	"math"

	"github.com/therobertcrocker/gm-toolkit/internal/spatial"
)

// Cube-coordinate hex line drawing, used to test warp line-of-sight. See
// https://www.redblobgames.com/grids/hexagons/#line-drawing.

type cubeCoord struct{ x, y, z int }

type fcube struct{ x, y, z float64 }

func axialToCube(h spatial.HexCoord) cubeCoord {
	return cubeCoord{x: h.Q, y: -h.Q - h.R, z: h.R}
}

func cubeToAxial(c cubeCoord) spatial.HexCoord {
	return spatial.HexCoord{Q: c.x, R: c.z}
}

func cubeDistance(a, b cubeCoord) int {
	return (absInt(a.x-b.x) + absInt(a.y-b.y) + absInt(a.z-b.z)) / 2
}

func hexDistance(a, b spatial.HexCoord) int {
	return cubeDistance(axialToCube(a), axialToCube(b))
}

func cubeRound(c fcube) cubeCoord {
	rx := math.Round(c.x)
	ry := math.Round(c.y)
	rz := math.Round(c.z)
	dx := math.Abs(rx - c.x)
	dy := math.Abs(ry - c.y)
	dz := math.Abs(rz - c.z)
	switch {
	case dx > dy && dx > dz:
		rx = -ry - rz
	case dy > dz:
		ry = -rx - rz
	default:
		rz = -rx - ry
	}
	return cubeCoord{x: int(rx), y: int(ry), z: int(rz)}
}

// hexLine returns the hexes the straight line between a and b passes through,
// inclusive of both endpoints. Endpoints are nudged by a sum-zero epsilon so a
// line crossing exactly on a hex border resolves deterministically.
func hexLine(a, b spatial.HexCoord) []spatial.HexCoord {
	ac := axialToCube(a)
	bc := axialToCube(b)
	n := cubeDistance(ac, bc)
	if n == 0 {
		return []spatial.HexCoord{a}
	}
	aN := fcube{float64(ac.x) + 1e-6, float64(ac.y) + 2e-6, float64(ac.z) - 3e-6}
	bN := fcube{float64(bc.x) + 1e-6, float64(bc.y) + 2e-6, float64(bc.z) - 3e-6}
	out := make([]spatial.HexCoord, 0, n+1)
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		out = append(out, cubeToAxial(cubeRound(fcube{
			x: lerp(aN.x, bN.x, t),
			y: lerp(aN.y, bN.y, t),
			z: lerp(aN.z, bN.z, t),
		})))
	}
	return out
}

func lerp(a, b, t float64) float64 { return a + (b-a)*t }

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
