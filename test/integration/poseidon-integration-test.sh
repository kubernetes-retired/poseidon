. common.sh

bash -x ./defaultns.sh
sleepNsecs 5

bash -x ./create-namespace.sh; 
sleepNsecs 5

bash -x ./create-deployment.sh; 
sleepNsecs 5

bash -x ./delete-deployment.sh; 
sleepNsecs 5

bash -x ./create-replicaset.sh;
sleepNsecs 5

bash -x ./update-replicaset.sh;
sleepNsecs 5

bash -x ./delete-replicaset.sh;
sleepNsecs 5

bash -x ./create-job.sh;
sleepNsecs 5

bash -x ./delete-job.sh; 
sleepNsecs 5

bash -x ./create-daemon.sh;
sleepNsecs 5

bash -x ./delete-daemon.sh;
sleepNsecs 5

bash -x ./create-statefulsets.sh;
sleepNsecs 5

bash -x ./delete-statefulsets.sh;
sleepNsecs 5

bash -x ./delete-namespace.sh; 
sleepNsecs 5

