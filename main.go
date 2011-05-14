package main

import (
//    "fmt"
    "flag"
    "image"
    "image/png"
    _ "image/jpeg"
    "math"
    "log"
    "os"
)

type FloatColor struct {
    R, G, B float32
}

type RingVectorRing struct {
    Size int
    DataCount int
    Data []image.RGBAColor
}

type RingVector struct {
    Size    int
    SizeInc int
    Count   int
    Rings []RingVectorRing
}

type RingVectorProperties struct {
    Loc image.Point
    RingCount int
    RingSize int
    RingSizeInc int
}

var (
    inputName   = flag.String("in", "", "input image")
    outputName  = flag.String("out", "", "output png")
    vectorX     = flag.Int("X", 30, "vector X location")
    vectorY     = flag.Int("Y", 30, "vector Y location")
    vectorSize  = flag.Int("S",  4, "vector radius")
    vectorRings = flag.Int("R", 1, "vector rings")
    ringSizeInc = flag.Int("I", 2, "ring size increment")
    threshold   = flag.Float64("T", 0.4, "threshold for drawing")
    skip        = flag.Int("K", 1, "skip of matching")
)

type LocationArray []image.Point
type IntLocationArrayMap map[int]LocationArray

type Bresenham struct {
    PixelCount map[int]int
    Pixels IntLocationArrayMap
}

func NewBresenham() *Bresenham {
    return &Bresenham{make(map[int]int), make(IntLocationArrayMap)}
}

var (
    bresenham *Bresenham
    BresenhamCirclePixelCount map[int]int
    BresenhamCirclePixels IntLocationArrayMap
)

type SamplingFunc func(x int, y int, idx int)
type CircleFunc func(x int, y int, idx int)

func SamplingRing(size int, t int, f SamplingFunc) {
    var x, y, d int 
    var t4, t3, t2 int
    x = 0
    y = size
    d = 3 - 2*size
    t2 = t / 2
    t3 = t * 3 / 4
    t4 = t / 4
    for x <= y {
        f( x, y, x)           // x
        f( y, x, t4 - x - 1)  // total / 4 - x
        f(-y, x, t4 + x)      // total / 4 + x
        f( x,-y, t2 - x - 1)  // total / 2 - x
        f(-x,-y, t2 + x)      // total / 2 + x
        f(-y,-x, t3 - x - 1)  // total 3 / 4 - x
        f( y,-x, t3 + x)      // total 3 / 4 + x
        f(-x, y, t  - x - 1)  // total - x
        
        if d <= 0 {
            d += 4 * x + 6
        } else {
            d += 4 * (x - y) + 10
            y -= 1
        }
        x += 1
    }
}

func (b *Bresenham) CalculateRing(size int){
    pxls := 0
    SamplingRing(size, 0, func(x int, y int, idx int){ pxls += 1 });
    b.PixelCount[size] = pxls
    
    arr := make(LocationArray, pxls)
    SamplingRing(size, pxls, func(x int, y int, idx int){ arr[idx] = image.Point{x,y} })
    b.Pixels[size] = arr
}

func (b *Bresenham) GetRing(size int)(*LocationArray, int) {
    if count, ok := b.PixelCount[size]; ok {
        arr := b.Pixels[size]
        return &arr, count
    }

    b.CalculateRing(size)
    count := b.PixelCount[size]
    arr   := b.Pixels[size]
    
    return &arr, count
}

func (b *Bresenham) GetPixelCount(size int)(int) {
    _, count := b.GetRing(size)
    return count
}

func (b *Bresenham) Run(size int, f CircleFunc){
    pxls, count := b.GetRing(size)
    for i := 0; i < count; i += 8 {
        f((*pxls)[i  ].X, (*pxls)[i  ].Y, i  )
        f((*pxls)[i+1].X, (*pxls)[i+1].Y, i+1)
        f((*pxls)[i+2].X, (*pxls)[i+2].Y, i+2)
        f((*pxls)[i+3].X, (*pxls)[i+3].Y, i+3)
        f((*pxls)[i+4].X, (*pxls)[i+4].Y, i+4)
        f((*pxls)[i+5].X, (*pxls)[i+5].Y, i+5)
        f((*pxls)[i+6].X, (*pxls)[i+6].Y, i+6)
        f((*pxls)[i+7].X, (*pxls)[i+7].Y, i+7)                                                        
    }
}

func CalcBresenhamCirclePixelCountMap(from int, to int)(*map[int]int){
    bh := make(map[int]int)
    for size := from; size <= to ; size += 1 {
        pxls := 0
        SamplingRing(size, 0, func(x int, y int, idx int){ pxls += 1 });
        bh[size] = pxls
    }
    return &bh
}

func GetRing(input *image.Image, size int, loc image.Point) (ring *RingVectorRing) {
    r := RingVectorRing{Size : size, DataCount: bresenham.GetPixelCount(size)}
    r.Data = make([]image.RGBAColor, r.DataCount )   
    bresenham.Run(size, func(x int, y int, idx int){
        r.Data[idx] = (*input).At(loc.X + x, loc.Y + y).(image.RGBAColor)
    })
    return &r
}

func Abs(val uint8) float64 {
    if val > 0 { 
        return float64(val)
    }
    return -float64(val)
}

func Diff(a uint8, b uint8) float64{
    if a > b {
        return float64(a - b)
    }
    return float64(b - a)
}

func ColorDiff(a image.RGBAColor, b image.RGBAColor) float64{
    return (Diff(a.R, b.R) +
            Diff(a.G, b.G) +
            Diff(a.B, b.B)) / 765.0;
}

func ColorSquareDiff(a image.RGBAColor, b image.RGBAColor) float64{
    return math.Sqrt( 
            math.Pow(float64(a.R) - float64(b.R), 2.0 ) +
            math.Pow(float64(a.R) - float64(b.R), 2.0 ) +
            math.Pow(float64(a.R) - float64(b.R), 2.0 )) / 765.0;
}

func MatchRing(sample *RingVector, ring *RingVector) (best float64) {
    best = 0.0
    for ringIdx := 0; ringIdx < sample.Count ; ringIdx += 1 {
        sampleRing := &sample.Rings[ringIdx]
        otherRing  := &ring.Rings[ringIdx]
        dc := sampleRing.DataCount
        bestRingMatch := 10000.0
        for i := 0; i < dc; i += *skip {
            curRingMatch := 0.0
            for j := 0; j < dc; j ++ {
                a := sampleRing.Data[j]
                b := otherRing.Data[(i+j)%dc]
                //curRingMatch += float64(Abs(int(a.R) - int(b.R)) + 
                //                 Abs(int(a.G) - int(b.G)) + 
                //                 Abs(int(a.B) - int(b.B))) / 768.0;
                curRingMatch += ColorDiff(a,b)
                if curRingMatch > bestRingMatch {
                    break
                }
            }
            if curRingMatch < bestRingMatch {
                bestRingMatch = curRingMatch
            }
            if bestRingMatch / float64(sampleRing.DataCount) < 0.01 {
                break
            }
        }
        best += bestRingMatch / float64(sampleRing.DataCount)
    }
    return
}

func GetRingVector(input *image.Image, vp RingVectorProperties)  *RingVector {
    v := RingVector{}
    v.Count = vp.RingCount
    v.Rings = make([]RingVectorRing, v.Count)
    v.SizeInc = vp.RingSizeInc
    size := vp.RingSize
    for i := 0; i < v.Count; i += 1 {
        v.Rings[i] = *GetRing(input, size, vp.Loc)
        size += v.SizeInc
    }
    return &v
}

func HighlightColor(a image.RGBAColor, h float64) image.RGBAColor{
    //if h < *threshold {
    //    return image.RGBAColor{0,0,0,255}
    //}
    //temp := math.Fmin((h - *threshold) / (1.0 - *threshold), 1.0)
    temp := math.Fmin(h, 1.0)
    c := uint8(temp * 255.0)
    return image.RGBAColor{c,c,c,255}
    //return image.RGBAColor{R:a.R, G:a.G, B:a.B, A:255)}
}

func MakeHeatMap(input *image.Image, output *image.RGBA, sample *RingVector, vp RingVectorProperties){
    largest := vp.RingSize + vp.RingSizeInc * vp.RingCount + 1
    dy := output.Bounds().Dy() - largest
    dx := output.Bounds().Dx() - largest
    for y := largest; y < dy; y++ { 
        for x := largest; x < dx; x++ { 
            vp.Loc.X = x
            vp.Loc.Y = y
            r := GetRingVector(input, vp)
            h := 1.0-MatchRing(sample, r)
            
            c := (*input).At(x, y).(image.RGBAColor)
            output.Set(x,y, HighlightColor(c, h))
        }
    }    
}

func run(input *image.Image, output *image.RGBA, vp RingVectorProperties){
    log.Println("Using vector : ", vp)

    dy := output.Bounds().Dy()
    dx := output.Bounds().Dx()
    for y := 0; y < dy; y++ { 
        for x := 0; x < dx; x++ { 
            output.Set(x,y, image.NRGBAColor{255,255,255, 255})
        }
    }
    sample := GetRingVector(input, vp)    
    loc := vp.Loc
    
    log.Println(sample)
    MakeHeatMap(input, output, sample, vp)
    
    for i := 0; i < sample.Count; i++ {
        bresenham.Run(sample.Rings[i].Size, func(x int, y int, idx int){
            output.Set(x + loc.X, y + loc.Y, image.RGBAColor{255,0,0,255})
        })
    }
}


func main() {
    
    flag.Parse()
    
    if *inputName == "" {
        log.Fatalln("No input defined");
    }    
    if *outputName == "" {
           *outputName = *inputName + ".heat.png"    
           log.Println("No output defined. Using " + *outputName + " instead.");
    }
    
    // open input file
    input, err := os.Open(*inputName, os.O_RDONLY, 0666)
    if err != nil {
        log.Fatalln(err) 
    }
    defer input.Close()
    
    // create output file
    output, err := os.Open(*outputName, os.O_CREATE | os.O_WRONLY, 0666)    
    if err != nil { 
        log.Fatalln(err) 
    } 
    defer output.Close()
        
    // decode png image
    inputImage, _, err := image.Decode(input)
    if err != nil {
        log.Fatalln(err)
    }
    
    outputImage := image.NewRGBA( inputImage.Bounds().Dx(), inputImage.Bounds().Dy() );
    
    vectorProperties := RingVectorProperties{ 
        Loc  : image.Point{*vectorX,*vectorY}, 
        RingSize : *vectorSize,
        RingCount : *vectorRings,
        RingSizeInc : *ringSizeInc}
      
    bresenham = NewBresenham()
    //BresenhamCirclePixelCount = *CalcBresenhamCirclePixelCountMap(1,50)
    run(&inputImage, outputImage, vectorProperties)
    
    if err = png.Encode(output, outputImage); err != nil {
        log.Fatalln(err) 
    } 
}
