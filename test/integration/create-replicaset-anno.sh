. common.sh

file_name="${manifests_root}/rs-annotations.yaml"
kind="replicaset"
object_name="nginx-anno"

create_object  $kind $object_name $file_name
