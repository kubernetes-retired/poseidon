[[ ${GUARDYVAR:-} -eq 1 ]] && return || readonly GUARDYVAR=1

RED='\033[0;31m  FAILED!!! '
YELLOW='\033[0;33m'
NC='\033[0m'

START=1
END=30
SRC=$(pwd)

KUBE_ROOT=$(readlink -m $(dirname "${BASH_SOURCE}")/../../)

manifests_root="${KUBE_ROOT}/test/integration/sample-yamls"

kubectl_cmd_path="${KUBE_ROOT}/test/integration/kubectl"
only_kubectl="${kubectl_cmd_path}/kubectl"

if [[ ! -f ${only_kubectl} ]]; then
  echo "${RED} $only_kubectl not available! ${NC}"
  exit 1
fi

host_kubectl="$only_kubectl  -s http://127.0.0.1:8080"

create_object () {
    echo "Create $1 $2 defined in file $3 :"
    cat $3
    echo "Start ... $(date)"
    SECONDS=0
    $host_kubectl create -f $3
    echo "Waiting for $1 $2 to be created..."

    for i in $(eval echo "{$START..$END}");do
        $host_kubectl get $1 $2 -o=jsonpath --template '{.metadata.name}';echo
        name="$($host_kubectl get -o=jsonpath $1 $2 --template '{.metadata.name}')"
        if [[ "$name" == "$2" ]];then
            echo "$1 $2 created:"
            $host_kubectl get $1
            break
        fi
        if [[ $i -eq $END ]];then
            echo -e "${RED} $1 $2 not created! giving up after $SECONDS seconds.${NC}"
            return 1;
        fi
        printf "."
        sleep 4
    done
    echo
    return 0
}

delete_object () {
    echo "list of existing $1 :"
    $host_kubectl get $1

    echo "$host_kubectl delete $1 $2 ----$(date)"
    SECONDS=0

    $host_kubectl delete $1 $2
    echo "Waiting for $1 $2 to be deleted..."
    for i in $(eval echo "{$START..$END}");do
        name="$($host_kubectl get $1 $2)"
        if [[ "$name" == "" ]];then
            echo "$1 $2 is deleted after $SECONDS second!"
            break
        fi
        if [[ $i -eq $END ]];then
            echo -e "${RED} $1 $2 is not deleted! giving up after $SECONDS seconds.${NC}"
            kube_ip_address=$cluster
            echo "Connect to the member cluster ($kube_ip_address) to check existence of $1 $2 :"
            $only_kubectl -s $kube_ip_address get $1 $2
            return 1
        fi
        printf "."
        sleep 2
     done
     echo

     echo "list of remaining $1 after deletion:"
     $host_kubectl get $1
     return 0
}

sleepNsecs() {
  echo ----------------------------------------------------------;echo
  echo " sleeping for $1 secs "
  i="0"
  while [ $i -lt $1 ]; do
    printf "."
    sleep 1
    i=$[ $i+1 ]
  done
  echo

}
