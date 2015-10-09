SWIG := swig3.0
SWIGGOFLAGS := -go -intgosize 64 
SWIGFLAGS := -c++ -I../firmament $(SWIGGOFLAGS)

all: firmament_wrap.cxx

firmament_wrap.cxx: firmament.i
	$(SWIG) $(SWIGFLAGS) $<

clean:
	rm *.cxx
	rm *.c
	rm *.go
