. common.sh

file_name="${manifests_root}/deployment.yaml"
kind="deployment"
object_name="nginx-deployment"

delete_object  $kind $object_name $file_name



file_name="${manifests_root}/deployment-annotation.yaml"
kind="deployment"
object_name="nginx-deployment-anno"

delete_object  $kind $object_name $file_name

