result=$(curl -X GET --header "Accept: */*" "https://cdc-rs-sftp-internal.azurewebsites.net/health")
echo "Response from server"
echo $result
exit
