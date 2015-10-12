EXTPATH := ../../ext

SWIG := swig3.0
SWIGGOFLAGS := -go -cgo -intgosize 64
SWIGFLAGS := -c++ -I../firmament $(SWIGGOFLAGS)

CXX = clang++
CXXFLAGS = -fPIC -I$(EXTPATH)/firmament/src

all: firmament_wrap.o

generate: firmament_wrap.cxx

%_wrap.cxx: %.i
	$(SWIG) $(SWIGFLAGS) $<

%.o: %.cxx
	$(CXX) $(CXXFLAGS) -c $<

#clang++ -c -fpic -I/home/malte/Projects/firmament/src/ -I/home/malte/Projects/firmament/ext/thread-safe-stl-containers-svn/ -I/home/malte/Projects/firmament/src/generated-cxx/ -std=c++11 -D__PLATFORM_HAS_BOOST__

clean:
	rm *.cxx
	rm *.c
	rm *.go
