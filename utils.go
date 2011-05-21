package main

import (
    "exp/draw"
    "image"
    _ "image/png"
    _ "image/jpeg"
)


func rgba(m image.Image) *image.RGBA {
    if r, ok := m.(*image.RGBA); ok {
        return r
    }
    b := m.Bounds()
    r := image.NewRGBA(b.Dx(), b.Dy())
    draw.Draw(r, b, m, image.ZP)
    return r
}
