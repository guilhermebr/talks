export PROMETHEUS_POD=$(kubectl get pods --namespace default -l "app=prometheus,component=server" -o jsonpath="{.items[0].metadata.name}")
export APP_POD=$(kubectl get pods --namespace default -l "app=app-resize" -o jsonpath="{.items[0].metadata.name}")
