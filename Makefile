all: server sivq

server:
	make -f Makefile.server

sivq:
	make -f Makefile.sivq

clean:
	make -f Makefile.server clean
	make -f Makefile.sivq clean
