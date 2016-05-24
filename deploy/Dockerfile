FROM ubuntu:14.04
MAINTAINER malte@csail.mit.edu, ionel.gog@cl.cam.ac.uk

RUN apt-get update \
    && apt-get -y install \
      libboost-system1.54 \
      libboost-chrono1.54 \
      libboost-thread1.54 \
      libboost-regex1.54 \
      libboost-timer1.54 \
      libhwloc5 \
      libgflags2 \
      libgoogle-glog0 \
      libssl1.0.0 \
      libctemplate2 \
      libprotobuf8 \
    && apt-get clean

RUN mkdir -p /usr/lib
COPY libcpprest.so.2.7 /usr/lib/
COPY libfirmament_scheduling.so /usr/lib/
COPY cs2.exe /usr/bin
COPY poseidon /usr/bin/

EXPOSE 8080
ENTRYPOINT ["/usr/bin/poseidon", "--logtostderr", "--cs2_binary=/usr/bin/cs2"]
