// 生成应用图标：蓝色驱动器（梯形机身 + 插槽条 + 指示灯），透明背景。
// 4 倍超采样后降采样，得到抗锯齿边缘，仅使用标准库。
//
// 用法：go run ./build/tools/icon <输出.png> [尺寸]
// 例：go run ./build/tools/icon build/appicon.png
//     go run ./build/tools/icon frontend/public/logo.png 256

// 4 倍超采样后降采样，得到抗锯齿边缘。仅使用标准库。
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const (
	size = 1024 // 输出尺寸
	ss   = 4    // 超采样倍数
)

type pt struct{ x, y, r float64 }

type rect struct{ x0, y0, x1, y1, r float64 }

var (
	cornerR = 100.0

	// 机身梯形（上窄下宽）与按 cornerR 内缩后的核心四边形
	core = []pt{
		{328.0, 220.0, 0},
		{696.0, 220.0, 0},
		{826.9, 800.0, 0},
		{197.1, 800.0, 0},
	}

	slot = rect{162, 688, 862, 864, 88}

	lamps = []pt{
		{657.6, 776, 41},
		{780.1, 776, 41},
	}

	drive = color.RGBA{14, 128, 200, 255} // #0E80C8
	white = color.RGBA{255, 255, 255, 255}
)

func insideRoundedRect(x, y float64, r rect) bool {
	if x < r.x0 || x > r.x1 || y < r.y0 || y > r.y1 {
		return false
	}
	if !((x < r.x0+r.r || x > r.x1-r.r) && (y < r.y0+r.r || y > r.y1-r.r)) {
		return true
	}
	corners := []pt{
		{r.x0 + r.r, r.y0 + r.r, r.r},
		{r.x1 - r.r, r.y0 + r.r, r.r},
		{r.x0 + r.r, r.y1 - r.r, r.r},
		{r.x1 - r.r, r.y1 - r.r, r.r},
	}
	for _, c := range corners {
		if math.Hypot(x-c.x, y-c.y) <= c.r {
			return true
		}
	}
	return false
}

func insideCircle(x, y float64, c pt) bool {
	return math.Hypot(x-c.x, y-c.y) <= c.r
}

// insideConvex 判断点是否在凸多边形内（边按顺时针给出）。
func insideConvex(x, y float64, poly []pt) bool {
	sign := 0.0
	for i := range poly {
		a := poly[i]
		b := poly[(i+1)%len(poly)]
		cross := (b.x-a.x)*(y-a.y) - (b.y-a.y)*(x-a.x)
		if cross != 0 {
			if sign == 0 {
				sign = math.Copysign(1, cross)
			} else if math.Copysign(1, cross) != sign {
				return false
			}
		}
	}
	return true
}

// distToSegment 点到线段的距离
func distToSegment(x, y float64, a, b pt) float64 {
	dx, dy := b.x-a.x, b.y-a.y
	len2 := dx*dx + dy*dy
	if len2 == 0 {
		return math.Hypot(x-a.x, y-a.y)
	}
	t := ((x-a.x)*dx + (y-a.y)*dy) / len2
	t = math.Max(0, math.Min(1, t))
	return math.Hypot(x-(a.x+t*dx), y-(a.y+t*dy))
}

// insideBody 机身 = 内缩多边形与半径 cornerR 圆盘的 Minkowski 和（即原梯形的圆角化）
func insideBody(x, y float64) bool {
	if insideConvex(x, y, core) {
		return true
	}
	for i := range core {
		if distToSegment(x, y, core[i], core[(i+1)%len(core)]) <= cornerR {
			return true
		}
	}
	return false
}

func render(x, y float64) color.RGBA {
	for _, l := range lamps {
		if insideCircle(x, y, l) {
			return drive
		}
	}
	if insideRoundedRect(x, y, slot) {
		return white
	}
	if insideBody(x, y) {
		return drive
	}
	return color.RGBA{0, 0, 0, 0}
}

func main() {
	out := os.Args[1]
	side := ""
	if len(os.Args) > 2 {
		side = os.Args[2]
	}

	hi := size * ss
	canvas := image.NewRGBA(image.Rect(0, 0, hi, hi))
	for py := 0; py < hi; py++ {
		for px := 0; px < hi; px++ {
			x := (float64(px) + 0.5) / ss
			y := (float64(py) + 0.5) / ss
			canvas.SetRGBA(px, py, render(x, y))
		}
	}

	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			var r, g, b, a float64
			for dy := 0; dy < ss; dy++ {
				for dx := 0; dx < ss; dx++ {
					c := canvas.RGBAAt(x*ss+dx, y*ss+dy)
					r += float64(c.R)
					g += float64(c.G)
					b += float64(c.B)
					a += float64(c.A)
				}
			}
			n := float64(ss * ss)
			img.SetRGBA(x, y, color.RGBA{uint8(r / n), uint8(g / n), uint8(b / n), uint8(a / n)})
		}
	}

	if side != "" {
		var target int
		fmt.Sscanf(side, "%d", &target)
		if target > 0 && target != size {
			scaled := image.NewRGBA(image.Rect(0, 0, target, target))
			factor := float64(size) / float64(target)
			for y := 0; y < target; y++ {
				for x := 0; x < target; x++ {
					scaled.SetRGBA(x, y, img.RGBAAt(int(float64(x)*factor), int(float64(y)*factor)))
				}
			}
			img = scaled
		}
	}

	f, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
	fmt.Printf("wrote %s (%dx%d)\n", out, img.Bounds().Dx(), img.Bounds().Dy())
}
