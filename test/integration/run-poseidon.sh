. common.sh

dir_test=./results

ls $dir_test -al || {
    echo -e "${RED} Make sure /var/run/kubernetes exists and is accessible to $USER:$USER ${NC}"
    exit 1
}

touch $dir_test/test123456 || {
    echo -e "${RED} Please run 'sudo chown -R $USER:$USER $dir_test' to change its owner to $USER! ${NC}"
    exit 1
}

ret="test$(date +%Y%m%d%H%M%S).txt"
bash -x ./poseidon-integration-test.sh  | tee $dir_test/$ret
