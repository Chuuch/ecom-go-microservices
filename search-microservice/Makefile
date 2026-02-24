docker-compose:
	docker compose -f docker-compose.yaml up --build

helm_install:
	helm install -f k8s/microservice/values.yaml search k8s/microservice

helm_uninstall:
	helm uninstall search

dry_run:
	helm template search k8s/microservice -f k8s/microservice/values.yaml

dry_run_install:
	helm install --dry-run --debug search k8s/microservice -f k8s/microservice/values.yaml

port_forward_microservice:
	kubectl port-forward service/microservice-service 8000:8000

port_forward_metrics:
	kubectl port-forward service/microservice-service 8001:8001

port_forward_rabbitmq:
	kubectl port-forward service/rabbitmq 15672:15672

port_forward_kibana:
	kubectl port-forward service/kibana 5601:5601

upload:
	docker build -t kuluruvineeth/simple-search-microservice:latest -f ./Dockerfile .
	docker push kuluruvineeth/simple-search-microservice:latest

port_forward_all:
	make port_forward_microservice &
	make port_forward_metrics &
	make port_forward_rabbitmq &
	make port_forward_kibana &
	wait

prometheus_install:
	helm install prometheus prometheus-community/kube-prometheus-stack

prometheus_uninstall:
	helm uninstall prometheus

minikube_start:
	minikube start --memory 7680 --cpus 4

# kill all port forwards
kill_port_forwards:
	@echo "Killing port forwards..."
	@lsof -ti:8000 | xargs kill 2>/dev/null || true
	@lsof -ti:8001 | xargs kill 2>/dev/null || true
	@lsof -ti:15672 | xargs kill 2>/dev/null || true
	@lsof -ti:5601 | xargs kill 2>/dev/null || true
	@pkill -f "kubectl port-forward" 2>/dev/null || true
	@sleep 1
	@echo "Port forwards killed"

install_all:
	helm install -f k8s/microservice/values.yaml search k8s/microservice
	kubectl create namespace monitoring
	helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring
	kubectl apply -f k8s/microservice/servicemonitor.yaml

monitoring_install:
	kubectl create namespace monitoring
	helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring
	kubectl apply -f k8s/microservice/servicemonitor.yaml

monitoring_uninstall:
	helm uninstall prometheus -n monitoring
	kubectl delete namespace monitoring

port_forward_monitoring:
	kubectl port-forward service/prometheus-prometheus-server 9090:9090