. common.sh

file_name="${manifests_root}/statefulsets.yaml"
kind="StatefulSet"
object_name="web"

delete_object  $kind $object_name $file_name

file_name="${manifests_root}/statefulsets-annotations.yaml"
kind="StatefulSet"
object_name="web-anno"

delete_object  $kind $object_name $file_name

