package main

import (
    "flag"
    "image"
    "image/png"
    _ "image/jpeg"
    "log"
    "os"
)

var (
    inputName   = flag.String("in", "", "input image")
    outputName  = flag.String("out", "", "output png")
    vectorX     = flag.Int("X", 272, "vector X location")
    vectorY     = flag.Int("Y", 274, "vector Y location")
    vectorSize  = flag.Int("S",  4, "vector radius")
    vectorRings = flag.Int("R", 1, "vector rings")
    ringSizeInc = flag.Int("I", 2, "ring size increment")
    threshold   = flag.Float64("T", 0.4, "threshold for drawing")
    rotStride   = flag.Float64("K", 0.001, "rotation stride")
    matchStride = flag.Int("M", 1, "matching value stride (can be set to 3 for grayscale pictures)")
    matchOffset = flag.Int("O", 0, "matching offset")
)

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
    input, err := os.OpenFile(*inputName, os.O_RDONLY, 0666)
    if err != nil {
        log.Fatalln(err) 
    }
    defer input.Close()
    
    // create output file
    output, err := os.OpenFile(*outputName, os.O_CREATE | os.O_WRONLY, 0666)    
    if err != nil { 
        log.Fatalln(err) 
    } 
    defer output.Close()
        
    // decode png image
    inputImage, _, err := image.Decode(input)
    if err != nil {
        log.Fatalln(err)
    }
    rgbaInput := rgba(inputImage)   
    
    vectorParams := RingVectorParameters{ 
        Radius : *vectorSize,
        Count : *vectorRings,
        RadiusInc : *ringSizeInc}
    
    sivqParams := SIVQParameters {
        RotationStride : float(*rotStride),
        MatchingStride : *matchStride,
        MatchingOffset : *matchOffset,
        Threshold : float(*threshold)}

    ringVector := NewRingVector(vectorParams)
    ringVector.LoadData(rgbaInput, *vectorX, *vectorY)

    outputImage := SIVQ(sivqParams, rgbaInput, ringVector)
    
    if err = png.Encode(output, outputImage); err != nil {
        log.Fatalln(err) 
    } 
}
