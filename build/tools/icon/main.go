// 应用图标工具（纯标准库）。
//
// 模式一：程序化渲染旧的蓝色驱动器图标（梯形机身 + 插槽条 + 指示灯），
// 4 倍超采样后降采样得到抗锯齿边缘。当前品牌图标已改为 build/appicon.svg
//（内嵌位图），此模式仅作历史参考。
//
//	用法：go run ./build/tools/icon <输出.png> [尺寸]
//	例：go run ./build/tools/icon build/appicon.png
//
// 模式二：把现有 PNG 按透明通道裁剪到内容包围盒，再等比缩放到指定尺寸（托盘图标）。
//
//	用法：go run ./build/tools/icon tray <输入.png> <输出.png> <尺寸>
//	例：go run ./build/tools/icon tray build/appicon.png build/windows/tray.png 32

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

// alphaBBox 返回非透明像素（alpha > 8）的包围盒。
func alphaBBox(img *image.RGBA) image.Rectangle {
	b := img.Bounds()
	minX, minY, maxX, maxY := b.Dx(), b.Dy(), -1, -1
	for y := 0; y < b.Dy(); y++ {
		for x := 0; x < b.Dx(); x++ {
			if img.RGBAAt(x, y).A > 8 {
				if x < minX {
					minX = x
				}
				if y < minY {
					minY = y
				}
				if x > maxX {
					maxX = x
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if maxX < 0 {
		return image.Rect(0, 0, 0, 0)
	}
	return image.Rect(minX, minY, maxX+1, maxY+1)
}

// boxDownscale 面积平均（box filter）降采样，factor 必须 >= 1 且为整数比例关系由调用方保证。
func boxDownscale(img *image.RGBA, w, h int) *image.RGBA {
	src := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	fx := float64(src.Dx()) / float64(w)
	fy := float64(src.Dy()) / float64(h)
	for y := 0; y < h; y++ {
		y0 := int(math.Floor(float64(y) * fy))
		y1 := int(math.Floor(float64(y+1) * fy))
		if y1 <= y0 {
			y1 = y0 + 1
		}
		for x := 0; x < w; x++ {
			x0 := int(math.Floor(float64(x) * fx))
			x1 := int(math.Floor(float64(x+1) * fx))
			if x1 <= x0 {
				x1 = x0 + 1
			}
			var r, g, b, a, n float64
			for sy := y0; sy < y1 && sy < src.Dy(); sy++ {
				for sx := x0; sx < x1 && sx < src.Dx(); sx++ {
					c := img.RGBAAt(sx, sy)
					r += float64(c.R)
					g += float64(c.G)
					b += float64(c.B)
					a += float64(c.A)
					n++
				}
			}
			if n > 0 {
				dst.SetRGBA(x, y, color.RGBA{uint8(r / n), uint8(g / n), uint8(b / n), uint8(a / n)})
			}
		}
	}
	return dst
}

// trayMode 从现有 PNG 生成托盘图标：裁剪到内容包围盒，等比缩放到边长 side，
// 居中放在透明画布上（留 1px 边距）。
func trayMode(args []string) {
	if len(args) != 3 {
		fmt.Fprintln(os.Stderr, "用法: icon tray <输入.png> <输出.png> <尺寸>")
		os.Exit(1)
	}
	in, out, sideStr := args[0], args[1], args[2]
	var side int
	if _, err := fmt.Sscanf(sideStr, "%d", &side); err != nil || side <= 0 {
		fmt.Fprintln(os.Stderr, "无效尺寸:", sideStr)
		os.Exit(1)
	}

	f, err := os.Open(in)
	if err != nil {
		panic(err)
	}
	src0, err := png.Decode(f)
	f.Close()
	if err != nil {
		panic(err)
	}
	src := image.NewRGBA(src0.Bounds())
	for y := src0.Bounds().Min.Y; y < src0.Bounds().Max.Y; y++ {
		for x := src0.Bounds().Min.X; x < src0.Bounds().Max.X; x++ {
			src.Set(x, y, src0.At(x, y))
		}
	}

	bbox := alphaBBox(src)
	if bbox.Empty() {
		fmt.Fprintln(os.Stderr, "输入图片全透明")
		os.Exit(1)
	}
	crop := src.SubImage(bbox).(*image.RGBA)

	fit := side - 2
	w, h := fit, fit
	if crop.Bounds().Dx() > crop.Bounds().Dy() {
		h = fit * crop.Bounds().Dy() / crop.Bounds().Dx()
	} else {
		w = fit * crop.Bounds().Dx() / crop.Bounds().Dy()
	}
	scaled := boxDownscale(crop, w, h)

	canvas := image.NewRGBA(image.Rect(0, 0, side, side))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			canvas.SetRGBA((side-w)/2+x, (side-h)/2+y, scaled.RGBAAt(x, y))
		}
	}

	o, err := os.Create(out)
	if err != nil {
		panic(err)
	}
	defer o.Close()
	if err := png.Encode(o, canvas); err != nil {
		panic(err)
	}
	fmt.Printf("wrote %s (%dx%d)\n", out, side, side)
}

func main() {
	if os.Args[1] == "tray" {
		trayMode(os.Args[2:])
		return
	}
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
