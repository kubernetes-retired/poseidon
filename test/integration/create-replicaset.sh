. common.sh

file_name="${manifests_root}/rs.yaml"
kind="replicaset"
object_name="nginx"

SECONDS=0
create_object  $kind $object_name $file_name

#add extra logic to make sure the newly-created replicaset is ready
DESIRED="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.spec.replicas}')"
echo
for i in `seq 1 60`; do
    ret1="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.status.replicas}' 2>/dev/null)"
    if [ "$ret1" == "$DESIRED" ]; then
       ret2="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.status.readyReplicas}' 2>/dev/null)"
       if [ "$ret2" == "$ret1" ];then
           break
       fi
    fi
    sleep 3
    printf "."
done
echo
$host_kubectl get $kind $object_name

if [ "$i" == 60 ]; then
    kubectl get $kind $object_name
    echo -e "${RED} [FAIL] Check status of $kind $object_name after creation Time out and give up after $SECONDS seconds! ${NC}"
    exit 1
else
    echo -e "${GREEN} [PASS] $kind $object_name created and get ready successfully after $SECONDS seconds! ${NC}"
fi

#check one pod's name
pod0="$($host_kubectl get pods -o=jsonpath --template '{.items[0].metadata.name}')"
if [[ "$pod0" =~ ^"nginx-" ]]; then
    ACTUAL_SCHEDULER="$($host_kubectl get pods -o=jsonpath --template '{.items[0].spec.schedulerName}' 2>/dev/null)"
    DESIRED_SCHEDULER="poseidon-scheduler"
    if [ "$DESIRED_SCHEDULER" == "$ACTUAL_SCHEDULER" ]; then
      echo -e "${GREEN} [PASS] pod $pod0 is created successfully ${NC}"
    else
      echo -e "${RED} [FAIL] pod $pod0 is not scheduled with $DESIRED_SCHEDULER but with $ACTUAL_SCHEDULER! ${NC}"
      exit 1
    fi
else
    echo -e "${RED} [FAIL] No pod detected for $kind $object_name! ${NC}"
fi

