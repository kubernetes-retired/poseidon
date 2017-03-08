FROM ubuntu:16.04
MAINTAINER malte@csail.mit.edu, ionel.gog@cl.cam.ac.uk

RUN apt-get update \
    && apt-get -y install \
      libboost-system1.55 \
      libboost-chrono1.55 \
      libboost-thread1.55 \
      libboost-regex1.55 \
      libboost-timer1.55 \
      libhwloc5 \
      libgflags2v5 \
      libgoogle-glog0v5 \
      libssl1.0.0 \
      libctemplate2v5 \
      libprotobuf-dev \
      libboost-all-dev \
    && apt-get clean

RUN mkdir -p /usr/lib
COPY libcpprest.so.2.7 /usr/lib/
COPY libfirmament_scheduling.so /usr/lib/
COPY libboost_system.so.1.58.0 /usr/lib/
COPY cs2.exe /usr/bin/
COPY poseidon /usr/bin/
COPY poseidon.cfg /usr/bin

EXPOSE 8080
ENTRYPOINT ["/usr/bin/poseidon", "--flagfile=/usr/bin/poseidon.cfg", "--logtostderr", "--cs2_binary=/usr/bin/cs2", "--max_tasks_per_pu=10"]
