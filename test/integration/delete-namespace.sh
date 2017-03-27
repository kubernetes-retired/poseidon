. common.sh

file_name="${manifests_root}/ns.yaml"
kind="namespace"
object_name="firmament-namespace"

delete_object  $kind $object_name $file_name


