echo "invoking functions..."
echo "test" | faas-cli invoke nodeinfo
echo "test" | faas-cli invoke figlet
echo "test" | faas-cli invoke sleep

# set proper replicas number
echo "scaling functions..."
kubectl scale --replicas=1 -n openfaas-fn deployment nodeinfo
echo "function nodeinfo scaled to 1"
kubectl scale --replicas=0 -n openfaas-fn deployment figlet
echo "function figlet scaled to 0"

