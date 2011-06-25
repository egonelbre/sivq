package main

import (
    "image"
    "math"
    _ "log"
)

var (
    circle Circle
)

func init() {
    circle = NewBresenham()
}

type float float32

type RingVectorParameters struct {
    Radius    int // initial radius
    Count     int // how many rings
    RadiusInc int // how much ring changes size
}

type SIVQParameters struct {
    GammaAdjustment float // for making images darker
    AverageBias     float // for using average around instead of center
    RotationStride  float // for calculating all possible rotations
    MatchingStride  int   // for comparing less values
    MatchingOffset  int   // for using different colors as comparison
    Threshold       float // minimal value to be show on output
    ProgressCallback func(float)
    StopCh          chan bool
}

type RingVectorRing struct {
    Radius int
    Stride int
    Data   []float
}

type RingVector struct {
    MinRadius      int
    MaxRadius      int
    TotalDataCount int
    Rings          []RingVectorRing
}


func NewRing(radius int) *RingVectorRing {
    r := RingVectorRing{Radius: radius, Stride: 3}
    pixelCount := circle.GetPixelCount(radius)
    r.Data = make([]float, pixelCount*3)
    return &r
}

func (r *RingVectorRing) LoadData(input *image.RGBA, X int, Y int) {
    inputStride := (*input).Stride
    pxls, count := circle.GetRing(r.Radius)
    i2 := 0
    for i := 0; i < count; i += 1 {
        x := (*pxls)[i].X
        y := (*pxls)[i].Y
        pixel := (*input).Pix[(Y+y)*inputStride+(X+x)]
        r.Data[i2+0] = float(pixel.R) / 255.0
        r.Data[i2+1] = float(pixel.G) / 255.0
        r.Data[i2+2] = float(pixel.B) / 255.0
        i2 += r.Stride
    }
}


func (r *RingVectorRing) LoadDataGray(input *FloatGray, X int, Y int) {
    inputStride := (*input).Stride
    circle.Run(r.Radius, func(x int, y int, idx int) {
        pixel := (*input).Pix[(Y+y)*inputStride+(X+x)]
        i := idx * r.Stride
        r.Data[i] = float(pixel.Y)
        r.Data[i+1] = float(pixel.Y)
        r.Data[i+2] = float(pixel.Y)
    })
}

func NewRingVector(rvp RingVectorParameters) *RingVector {
    rv := RingVector{}
    rv.Rings = make([]RingVectorRing, rvp.Count)
    rv.TotalDataCount = 0
    radius := rvp.Radius
    rv.MinRadius = radius
    for i := range rv.Rings {
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
    for i := range rv.Rings {
        nrv.Rings[i] = *NewRing(rv.Rings[i].Radius)
    }
    return &nrv
}


func (rv *RingVector) LoadData(input *image.RGBA, X int, Y int) {
    for _, r := range rv.Rings {
        r.LoadData(input, X, Y)
    }
}

func (rv *RingVector) LoadDataGray(input *FloatGray, X int, Y int) {
    for _, r := range rv.Rings {
        r.LoadDataGray(input, X, Y)
    }
}

func (rv *RingVector) Average() float {
    var avg float
    avg = 0.0
    count := 0
    for _, r := range rv.Rings {
        for _, val := range r.Data {
           count += 1
           avg += val
        }
    }
    avg = avg / float(count)
    return avg
}

type RingDiff struct {
    Base      int
    Diff      float
    DiffCount int
}

func (A *RingVector) Diff(B *RingVector, p SIVQParameters) (best float) {
    best = float(math.Inf(1))

    cache := make([]*RingDiff, len(A.Rings))
    for ri := range A.Rings {
        cache[ri] = &RingDiff{Base: -1}
    }

    for rotation := float(0.0); rotation < Tau; rotation += p.RotationStride {
        total := float(0.0)
        totalCount := 0
        for ri := range A.Rings {
            dA := A.Rings[ri].Data
            dB := B.Rings[ri].Data

            stride := A.Rings[ri].Stride
            dataCount := len(dA)

            diff := float(0.0)
            diffCount := 0

            base := int((rotation / Tau) * float(dataCount))
            base = base - base%stride

            cacheVal := cache[ri]
            if cacheVal.Base != base {
                i2 := (base + p.MatchingOffset) % dataCount
                for i := p.MatchingOffset; i < dataCount; i += p.MatchingStride {
                    d := dA[i] - dB[i2]
                    diff += d * d
                    diffCount += 1
                    i2 += p.MatchingStride
                    if i2 >= dataCount {
                        i2 = i2 % dataCount
                    }
                }
                cacheVal.Base = base
                cacheVal.Diff = diff
                cacheVal.DiffCount = diffCount
            } else {
                diff = cacheVal.Diff
                diffCount = cacheVal.DiffCount
            }

            total += diff
            totalCount += diffCount
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

func calculateSIVQ(p SIVQParameters, input *image.RGBA, output *FloatGray, rv *RingVector){
    startAtX := rv.MaxRadius
    startAtY := rv.MaxRadius
    stopAtX := output.Bounds().Dx() - rv.MaxRadius
    stopAtY := output.Bounds().Dy() - rv.MaxRadius
    
    cancel := make(chan int)
    done := make(chan int)

    routineCount := 0

    for y := startAtY; y < stopAtY; y++ {
        routineCount += 1
        go func(y int) {
            r := rv.EmptyClone()
            for x := startAtX; x < stopAtX; x++ {
                r.LoadData(input, x, y)
                output.Set(x, y, FloatGrayColor{rv.Diff(r, p)})
                select {
                case <-cancel:
                    break
                default:
                }
            }
            done <- y
        }(y)
    }
    
    for i := 0; i < routineCount; i++ {
        select {
        case y := <-done:
            p.ProgressCallback(float(y-startAtY) / float(stopAtY-startAtY-1))
        case <-p.StopCh:
            for j := i; j < routineCount; j++ {
                cancel <- 1
            }
        }
    }
}

func fixCircleDefects(p SIVQParameters, input *FloatGray, output *FloatGray, rv *RingVector) {
    startAtX := rv.MaxRadius
    startAtY := rv.MaxRadius
    stopAtX := output.Bounds().Dx() - rv.MaxRadius
    stopAtY := output.Bounds().Dy() - rv.MaxRadius

    cancel := make(chan int)
    done := make(chan int)

    routineCount := 0

    for y := startAtY; y < stopAtY; y++ {
        routineCount += 1
        go func(y int) {
            r := rv.EmptyClone()
            for x := startAtX; x < stopAtX; x++ {
                r.LoadDataGray(input, x, y)
                inY := input.At(x,y).(FloatGrayColor).Y
                output.Set(x, y, 
                    FloatGrayColor{ (p.AverageBias * r.Average()) + (1.0 - p.AverageBias) * inY } )
            }
            done <- y
        }(y)
    }
    
    for i := 0; i < routineCount; i++ {
        select {
        case _ = <-done:
            //p.ProgressCallback(float(y-startAtY) / float(stopAtY-startAtY-1))
        case <-p.StopCh:
            for j := i; j < routineCount; j++ {
                cancel <- 1
            }
        }
    }
}

func SIVQ(p SIVQParameters, input *image.RGBA, rv *RingVector) *image.RGBA {
    if p.ProgressCallback == nil { 
        p.ProgressCallback = func(p float){}
    }
    if p.StopCh == nil {
        p.StopCh = make(chan bool)
    }
    
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
    
    if p.AverageBias > 1.0 {
        p.AverageBias = 1.0
    } else if p.AverageBias < 0.0 {
        p.AverageBias = 0.0
    }
    
    dx := input.Bounds().Dx()
    dy := input.Bounds().Dy()
    
    temp := NewFloatGray(dx, dy)
    for i := range temp.Pix {
        temp.Pix[i] = FloatGrayColor{0.0}
    }
    calculateSIVQ(p, input, temp, rv)
    var output * FloatGray
    if p.AverageBias >= 0.001 {
        output = NewFloatGray(dx, dy)
        fixCircleDefects(p, temp, output, rv)
    } else {
        output = temp
    }
    
    return output.ToRGBA(p.GammaAdjustment, p.Threshold)
}
