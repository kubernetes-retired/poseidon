EXTPATH := ../../ext

FIRMAMENT_INCLUDEPATHS = -I$(EXTPATH)/firmament/src \
												 -I$(EXTPATH)/firmament/ext/thread-safe-stl-containers-svn/ \
												 -I$(EXTPATH)/firmament/src/generated-cxx/

SWIG := swig3.0
SWIGGOFLAGS := -go -cgo -intgosize 64
SWIGFLAGS := -c++ -I../firmament $(SWIGGOFLAGS)

CXX = clang++
CXXFLAGS = -fPIC -std=c++ -D__PLATFORM_HAS_BOOST__ $(FIRMAMENT_INCLUDEPATHS)

all: firmament_wrap.o

generate: firmament_wrap.cxx

%_wrap.cxx: %.i
	$(SWIG) $(SWIGFLAGS) $<

%.o: %.cxx
	$(CXX) $(CXXFLAGS) -c $<

clean:
	rm *.cxx
	rm *.c
	rm *.go
