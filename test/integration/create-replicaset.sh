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
    echo -e "${RED}Check status of $kind $object_name after creation Time out and give up after $SECONDS seconds! ${NC}"
    exit 1
else
    echo "$kind $object_name created and get ready successfully after $SECONDS seconds!"
fi

#check one pod's name
pod0="$($host_kubectl get pods -o=jsonpath --template '{.items[0].metadata.name}')"
if [[ "$pod0" =~ ^"nginx-" ]]; then
    echo "pod $pod0 is created successfully"
else
    echo -e "${RED} No pod detected for $kind $object_name! ${NC}"
fi


