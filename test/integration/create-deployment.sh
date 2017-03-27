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
    echo -e "${RED} Check status of $kind $object_name Time out and give up after $SECONDS seconds! ${NC}"
    exit 1
else
    echo "$kind $object_name created successfully after $SECONDS seconds."
fi


