. common.sh

file_name="${manifests_root}/rs.yaml"
kind="replicaset"
object_name="nginx"

SECONDS=0
delete_object  $kind $object_name $file_name
if [[ $? -eq 0 ]]; then
    echo "$kind $object_name is deleted after $SECONDS seconds!"
fi


file_name="${manifests_root}/rs-annotations.yaml"
kind="replicaset"
object_name="nginx-anno"

SECONDS=0
delete_object  $kind $object_name $file_name
if [[ $? -eq 0 ]]; then
    echo "$kind $object_name is deleted after $SECONDS seconds!"
fi

