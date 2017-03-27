. common.sh

file_name="${manifests_root}/daemonset.yaml"
kind="DaemonSet"
object_name="nginx-daemon"

create_object  $kind $object_name $file_name

echo
for i in `seq 1 60`; do
    DESIRED="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.status.desiredNumberScheduled}' 2>/dev/null)"
    ret1="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.status.currentNumberScheduled}' 2>/dev/null)"
    if [ "$ret1" == "$DESIRED" ]; then
        ret2="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.status.numberReady}' 2>/dev/null)"
        if [ "$ret1" == "$ret2" ]; then
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

# Job annotation with poseidon-scheduler
file_name="${manifests_root}/daemonset-annotations.yaml"
kind="DaemonSet"
object_name="nginx-daemon-anno"

create_object  $kind $object_name $file_name

DESIRED="1"
echo
for i in `seq 1 60`; do
    DESIRED="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.status.desiredNumberScheduled}' 2>/dev/null)"
    ret1="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.status.currentNumberScheduled}' 2>/dev/null)"
    if [ "$ret1" == "$DESIRED" ]; then
        ret2="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.status.numberReady}' 2>/dev/null)"
        if [ "$ret1" == "$ret2" ]; then
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

