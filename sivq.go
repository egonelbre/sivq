package main

import (
    "image"
    "math"
)

var (
    circle Circle
)

func init(){
    circle = NewBresenham()
}

type float float32

type RingVectorParameters struct {
    Radius int     // initial radius
    Count int      // how many rings
    RadiusInc int  // how much ring changes size
}

type SIVQParameters struct {
    GammaAdjustment float // for making images darker
    RotationStride float // for calculating all possible rotations
    MatchingStride int   // for comparing less values
    MatchingOffset int   // for using different colors as comparison
    Threshold float      // minimal value to be show on output
}

type RingVectorRing struct {
    Radius int
    Stride int
    Data []float
}

type RingVector struct {
    MinRadius int
    MaxRadius int
    TotalDataCount int
    Rings []RingVectorRing
}


func NewRing(radius int) *RingVectorRing {
    r := RingVectorRing{ Radius : radius, Stride : 3 }
    pixelCount := circle.GetPixelCount( radius )
    r.Data = make([]float, pixelCount * 3)
    return &r
}

func (r *RingVectorRing) LoadData(input *image.RGBA, X int, Y int) {
    inputStride := (*input).Stride
    circle.Run(r.Radius, func(x int, y int, idx int){
        pixel := (*input).Pix[ (Y + y) * inputStride + (X + x) ]
        i := idx * r.Stride
        r.Data[i + 0] = float( pixel.R ) / 255.0
        r.Data[i + 1] = float( pixel.G ) / 255.0
        r.Data[i + 2] = float( pixel.B ) / 255.0
    })
}

func NewRingVector(rvp RingVectorParameters) *RingVector {
    rv := RingVector{}
    rv.Rings = make([]RingVectorRing, rvp.Count)    
    rv.TotalDataCount = 0
    radius := rvp.Radius
    rv.MinRadius = radius
    for i, _ := range rv.Rings {
        rv.Rings[i] = *NewRing(radius)
        rv.TotalDataCount += len(rv.Rings[i].Data)
        radius += rvp.RadiusInc
    }
    rv.MaxRadius = radius - rvp.RadiusInc
    return &rv
}

func (rv *RingVector) EmptyClone() *RingVector {
    nrv := RingVector{}
    nrv.MinRadius = rv.MinRadius
    nrv.MaxRadius = rv.MaxRadius
    nrv.TotalDataCount = rv.TotalDataCount
    nrv.Rings = make([]RingVectorRing, len(rv.Rings))
    for i, _ := range rv.Rings {
        nrv.Rings[i] = *NewRing( rv.Rings[i].Radius )
    }    
    return &nrv
}


func (rv *RingVector) LoadData(input *image.RGBA, X int, Y int){
    for _, r := range rv.Rings {
        r.LoadData(input, X, Y)
    }
}

func (A *RingVector) Diff(B *RingVector, p SIVQParameters) (best float){
    best = float(math.Inf(1))
    for rotation := float(0.0); rotation < Tau; rotation += p.RotationStride {
        total := float(0.0)
        totalCount := 0
        for ri, _ := range A.Rings {
            dA := &A.Rings[ri].Data
            dB := &B.Rings[ri].Data

            stride := A.Rings[ri].Stride
            dataCount := len(*dA)
            
            base := int((rotation / Tau) * float(dataCount))
            base = base - base % stride
                                    
            for i := p.MatchingOffset; i < dataCount; i += p.MatchingStride {
                d := (*dA)[i] - (*dB)[ (base + i) % dataCount ]
                total += d * d
                totalCount += 1
            }
        }
        total = total / float(totalCount)
        if best > total {
            best = total
        }
        if best <= 0.000025 {
            break
        }
    }
    best = float(math.Sqrt(float64(best)))
    return best
}

func highlightColor(p SIVQParameters, a image.RGBAColor, h float) image.RGBAColor{
    h = 1.0 - h
    // adjust gamma for clarity
    h = float(math.Pow(float64(h), float64(p.GammaAdjustment)))
    if h > 1.0 { h = 1.0 } else if h < 0.0 { h = 0.0 }
    if h < p.Threshold {
        return image.RGBAColor{0,0,0,255}
    }
    c := uint8(h * 255.0)
    return image.RGBAColor{c,c,c,255}
}

func makeHeatMap(p SIVQParameters, input *image.RGBA, output *image.RGBA, rv *RingVector){
    startAtX := rv.MaxRadius
    startAtY := rv.MaxRadius
    stopAtX  := output.Bounds().Dx() - rv.MaxRadius
    stopAtY  := output.Bounds().Dy() - rv.MaxRadius
    
    minStride := Tau
    for _, r := range rv.Rings {
        stride := Tau * float(r.Stride) / float(len(r.Data))
        if minStride > stride {
            minStride = stride
        }
    }
    if p.RotationStride < minStride {
        p.RotationStride = minStride
    }
    
    quit := make(chan int) // external
    cancel := make(chan int)
    done := make(chan int)
        
    routineCount := 0   
    
    for y := startAtY ; y < stopAtY; y++ {
        routineCount += 1
        go func(y int){
            r := rv.EmptyClone()
            for x := startAtX; x < stopAtX; x++ {
                r.LoadData(input, x, y)
                c := (*input).At(x, y).(image.RGBAColor)
                output.Set(x, y, highlightColor(p, c, rv.Diff(r, p)))
                select {
                    case <-cancel : break
                    default:
                }
            }
            done <- y
        }(y)
    }
    
    for i := 0; i < routineCount; i++ {
        select {
        case <-done:
        case <-quit:
            for j := i; j < routineCount; j++ {
                cancel <- 1
            }
        }
    }
}

func SIVQ(p SIVQParameters, input *image.RGBA, rv *RingVector) *image.RGBA {
    dx := input.Bounds().Dx()
    dy := input.Bounds().Dy()
    output := image.NewRGBA( dx, dy )
    for i, _ := range output.Pix {
        output.Pix[i] = image.RGBAColor{0,0,0,255}
    }
    makeHeatMap(p, input, output, rv)
    return output
}

