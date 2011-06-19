package main

import (
    "image"
    "math"
)


const (
    Tau = float(2 * math.Pi)
)

type Circle interface {
    Run(size int, f SamplingFunc)
    GetPixelCount(size int) int
}

type SamplingFunc func(x int, y int, idx int)

func SamplingRing(size int, count int, f SamplingFunc) {
    var x, y, d int
    var sec4, sec3, sec2, sec1 int
    x = 0
    y = size
    d = 3 - 2*size
    
    sec1 = count/4;
    sec2 = 2*sec1;
    sec3 = 3*sec1;
    sec4 = 4*sec1;
    
    for x <= y {
        f(x, y, x)
        if x != 0 { f(-x,  y, sec4 - x) }
        if y != 0 { f( x, -y, sec2 - x) }
        if (x != 0) && (y != 0) { f(-x, -y, sec2 + x) }
        if x != y {
            f(y, x, sec1 - x)
            if  x != 0 { f( y, -x, sec1 + x) }
            if  y != 0 { f(-y,  x, sec3 + x) }
            if  (x != 0) && (y != 0) { f(-y, -x, sec3 - x) }
        }
        
        if d <= 0 {
            d += 4*x + 6
        } else {
            d += 4*(x-y) + 10
            y -= 1
        }
        x += 1
    }
}

type LocationArray []image.Point
type IntLocationArrayMap map[int]LocationArray

type Bresenham struct {
    PixelCount map[int]int
    Pixels     IntLocationArrayMap
}

func NewBresenham() *Bresenham {
    return &Bresenham{make(map[int]int), make(IntLocationArrayMap)}
}

func (b *Bresenham) CalculateRing(radius int) {
    pxls := 0
    SamplingRing(radius, 0, func(x int, y int, idx int) { pxls += 1 })
    b.PixelCount[radius] = pxls

    arr := make(LocationArray, pxls)
    SamplingRing(radius, pxls,
        func(x int, y int, idx int) {
            arr[idx] = image.Point{x, y}
        })
    b.Pixels[radius] = arr
}

func (b *Bresenham) GetRing(radius int) (*LocationArray, int) {
    if count, ok := b.PixelCount[radius]; ok {
        arr := b.Pixels[radius]
        return &arr, count
    }

    b.CalculateRing(radius)
    count := b.PixelCount[radius]
    arr := b.Pixels[radius]

    return &arr, count
}

func (b *Bresenham) GetPixelCount(radius int) int {
    _, count := b.GetRing(radius)
    return count
}

func (b *Bresenham) Run(size int, f SamplingFunc) {
    pxls, count := b.GetRing(size)
    for i := 0; i < count; i += 4 {
        f((*pxls)[i].X, (*pxls)[i].Y, i)
        f((*pxls)[i+1].X, (*pxls)[i+1].Y, i+1)
        f((*pxls)[i+2].X, (*pxls)[i+2].Y, i+2)
        f((*pxls)[i+3].X, (*pxls)[i+3].Y, i+3)
    }
}
