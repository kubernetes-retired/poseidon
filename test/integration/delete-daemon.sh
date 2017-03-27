. common.sh

file_name="${manifests_root}/daemonset.yaml"
kind="DaemonSet"
object_name="nginx-daemon"

delete_object  $kind $object_name $file_name

file_name="${manifests_root}/daemonset-annotations.yaml"
kind="DaemonSet"
object_name="nginx-daemon-anno"

delete_object  $kind $object_name $file_name

