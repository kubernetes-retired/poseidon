. common.sh

echo "check default namespace status"

i="0"
while [[ $i -lt 60 ]]; do
   status="$($host_kubectl get -o=jsonpath namespace default --template '{.metadata.name}')"
   if [[ "$status" == "default" ]];then
        echo "default namespace exists!"
        break
   fi
   i=$[ $i+1 ]
   printf "."; sleep 2
done

$host_kubectl get namespace

if [[ $i == 60 ]];then
    echo -e "${RED} default namespace not exist. Please press ctrl-c to exit the test!${NC}"
    read -n 1;echo
    exit 1
fi


