package main

import (
    "image"
    "math"
)


type FloatGrayColor struct {
    Y float
}

func (c FloatGrayColor) RGBA() (r,g,b,a uint32) {
    var y uint32
    if c.Y > 1.0 {
        y = 0xffff
    } else if c.Y < 0.0 {
        y = 0
    } else {
        y = (uint32)(c.Y * 0xffff)
    }
    return y, y, y, 0xffff
}

func toFloatGrayColor(c image.Color) image.Color {
    if _, ok := c.(FloatGrayColor); ok {
        return c
    }
    r, g, b, _ := c.RGBA()
    y :=  (0.3*float(r) + 0.59*float(g) + 0.11*float(b)) / float(0xffff)
    return FloatGrayColor{ y }
}

var FloatGrayColorModel image.ColorModel = image.ColorModelFunc(toFloatGrayColor)

type FloatGray struct {
    Pix     []FloatGrayColor
    Stride  int
    Rect    image.Rectangle
}

func (p *FloatGray) ColorModel() image.ColorModel { return FloatGrayColorModel }

func (p *FloatGray) Bounds() image.Rectangle { return p.Rect }

func (p *FloatGray) At(x, y int) image.Color {
    if !p.Rect.Contains(image.Point{x, y}) {
        return FloatGrayColor{}
    }
    return p.Pix[y*p.Stride+x]
}

func (p *FloatGray) AtFast(x, y int) image.Color{
    return p.Pix[y*p.Stride+x]
}

func (p *FloatGray) Set(x, y int, c image.Color) {
    if !p.Rect.Contains(image.Point{x, y}) {
        return
    }
    p.Pix[y*p.Stride+x] = toFloatGrayColor(c).(FloatGrayColor)
}

func (p *FloatGray) SetFloatGray(x, y int, c FloatGrayColor) {
    if !p.Rect.Contains(image.Point{x, y}) {
        return
    }
    p.Pix[y*p.Stride+x] = c
}

// Opaque scans the entire image and returns whether or not it is fully opaque.
func (p *FloatGray) Opaque() bool {
    return true
}

func (p *FloatGray) ToRGBA(gamma float, threshold float) * image.RGBA {
    rgba := image.NewRGBA(p.Rect.Dx(), p.Rect.Dy())
    if threshold < 0.0 {
        threshold = 0.0
    }

    for i, c := range p.Pix {
        y := float(math.Pow(float64(1.0 - c.Y), float64(gamma)))
        
        if y < threshold {
            y = 0
        } else if y > 1.0 {
            y = 1.0
        }
        
        v := uint8(y * 255.0)
        rgba.Pix[i] = image.RGBAColor{v, v, v, 255}
    }
    return rgba
}

// NewGray16 returns a new Gray16 with the given width and height.
func NewFloatGray(w, h int) *FloatGray {
    pix := make([]FloatGrayColor, w*h)
    return &FloatGray{pix, w, image.Rectangle{image.ZP, image.Point{w, h}}}
}

