. common.sh

file_name="${manifests_root}/deployment.yaml"
kind="deployment"
object_name="nginx-deployment"
SECONDS=0
create_object  $kind $object_name $file_name

DESIRED="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.spec.replicas}')"
echo "Run these commands to check $kind $object_name status:"
echo "$host_kubectl get $kind $object_name -o=jsonpath --template '{.status.replicas}'"
echo "$host_kubectl get $kind $object_name -o=jsonpath --template '{.status.availableReplicas}'"

for i in `seq 1 60`; do
    ret1="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.status.replicas}' 2>/dev/null)"
    if [ "$ret1" == "$DESIRED" ]; then
       ret2="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.status.availableReplicas}' 2>/dev/null)"
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
    echo -e "${RED} [FAIL] Check status of $kind $object_name Time out and give up after $SECONDS seconds! ${NC}"
    exit 1
else
    echo -e "${GREEN} [PASS] $kind $object_name created successfully after $SECONDS seconds. ${NC}"
fi

file_name="${manifests_root}/deployment-annotations.yaml"
kind="deployment"
object_name="nginx-deployment-anno"
SECONDS=0
create_object  $kind $object_name $file_name

DESIRED="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.spec.replicas}')"
echo "Run these commands to check $kind $object_name status:"
echo "$host_kubectl get $kind $object_name -o=jsonpath --template '{.status.replicas}'"
echo "$host_kubectl get $kind $object_name -o=jsonpath --template '{.status.availableReplicas}'"

for i in `seq 1 60`; do
    ret1="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.status.replicas}' 2>/dev/null)"
    if [ "$ret1" == "$DESIRED" ]; then
       ret2="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.status.availableReplicas}' 2>/dev/null)"
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
    echo -e "${RED} [FAIL] Check status of $kind $object_name Time out and give up after $SECONDS seconds! ${NC}"
    exit 1
else
    ACTUAL_SCHEDULER="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.spec.template.spec.schedulerName}' 2>/dev/null)"
    DESIRED_SCHEDULER="poseidon-scheduler"
    if [ "$DESIRED_SCHEDULER" == "$ACTUAL_SCHEDULER" ]; then
       echo -e "${GREEN} [PASS] $kind $object_name created successfully after $SECONDS seconds. ${NC}"
    else
      echo -e "${RED} [FAIL] Check status of $kind $object_name not scheduled with $DESIRED_SCHEDULER but with $ACTUAL_SCHEDULER! ${NC}"
      exit 1
    fi
fi
