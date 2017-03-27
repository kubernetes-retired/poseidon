. common.sh

file_name="${manifests_root}/job.yaml"
kind="job"
object_name="hello"

delete_object  $kind $object_name $file_name

file_name="${manifests_root}/job-annotations.yaml"
kind="job"
object_name="cpuspin"

delete_object  $kind $object_name $file_name

