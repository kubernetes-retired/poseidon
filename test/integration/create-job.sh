. common.sh

file_name="${manifests_root}/job.yaml"
kind="job"
object_name="hello"

create_object  $kind $object_name $file_name

DESIRED="1"
echo
for i in `seq 1 60`; do
    ret1="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.status.succeeded}' 2>/dev/null)"
    if [ "$ret1" == "$DESIRED" ]; then
           break
    fi
    sleep 3
    printf "."
done
echo
$host_kubectl get $kind $object_name

if [ "$i" == 60 ]; then
    kubectl get $kind $object_name
    echo -e "${RED} Check status of $kind $object_name Time out and give up after $SECONDS seconds! ${NC}"
    exit 1
else
    echo "$kind $object_name created successfully after $SECONDS seconds."
fi

# Job annotation with poseidon-scheduler
file_name="${manifests_root}/job-annotations.yaml"
kind="job"
object_name="cpuspin"

create_object  $kind $object_name $file_name

DESIRED="1"
echo
for i in `seq 1 60`; do
    ret1="$($host_kubectl get $kind $object_name -o=jsonpath --template '{.status.succeeded}' 2>/dev/null)"
    if [ "$ret1" == "$DESIRED" ]; then
           break
    fi
    sleep 3
    printf "."
done
echo
$host_kubectl get $kind $object_name

if [ "$i" == 60 ]; then
    kubectl get $kind $object_name
    echo -e "${RED} Check status of $kind $object_name Time out and give up after $SECONDS seconds! ${NC}"
    exit 1
else
    echo "$kind $object_name created successfully after $SECONDS seconds."
fi

