all: server sivq

server: *.go
	make -f Makefile.server clean
	make -f Makefile.server

sivq: *.go
	make -f Makefile.sivq clean
	make -f Makefile.sivq

.PHONY: test clean

test:
	./sivq -in data/shapes.png -out test/shape-box-top.png   -X 81  -Y 47  -I 2 -S 4 -R 3 -M 1
	./sivq -in data/shapes.png -out test/shape-circle-NW.png -X 184 -Y 52  -I 2 -S 4 -R 3 -M 1
	./sivq -in data/letters.png -out test/letters-A.png      -X 27  -Y 35  -I 1 -S 4 -R 3 -M 1
	./sivq -in data/tumor.png -out test/tumor-pink.png       -X 460 -Y 170 -I 2 -S 5 -R 3 -M 6
	./sivq -in data/tumor2x.png -out test/tumor2x-pink.png   -X 920 -Y 340 -I 2 -S 5 -R 3 -M 6

clean:
	make -f Makefile.server clean
	make -f Makefile.sivq clean
