

build:
	cd src; go vet ./...
	cd src; go build ./...	

docker-build:
	docker build src/. -t ryannedolan/substation:dev

deploy: undeploy docker-build
	helm install substation deploy/helm/substation --atomic --set image.tag=dev

undeploy:
	@-helm history substation 2> /dev/null && helm uninstall substation || echo "no existing release."

target/kind-cluster.out:
	kind create cluster --config deploy/kind/cluster.yaml
	mkdir -p target
	touch target/kind-cluster.out

target/install-istio.out: target/istio-1.7.0/bin/istioctl
	target/istio-1.7.0/bin/istioctl install --set profile=demo
	kubectl apply -f target/istio-1.7.0/samples/addons -n istio-system || sleep 5 && kubectl apply -f target/istio-1.7.0/samples/addons/kiali.yaml -n istio-system
	touch target/install-istio.out

kind-load: docker-build target/kind-cluster.out
	kind load docker-image ryannedolan/substation:dev

kind-deploy: target/kind-cluster.out target/kind-setup.out target/install-istio.out kind-load deploy

target/istio-1.7.0/bin/istioctl:
	cd target; curl -L https://istio.io/downloadIstio | ISTIO_VERSION=1.7.0 sh -

target/kind-setup.out: target/kind-cluster.out
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.0.0/aio/deploy/recommended.yaml
	-kubectl create clusterrolebinding default-admin --clusterrole cluster-admin --serviceaccount=default:default
	mkdir -p target
	touch target/kind-setup.out

target/bearer-token: target/kind-setup.out
	mkdir -p target
	kubectl get secrets -o jsonpath="{.items[?(@.metadata.annotations['kubernetes\.io/service-account\.name']=='default')].data.token}"|base64 --decode > target/bearer-token

.PHONY: clean
clean:
ifneq (,$(wildcard target/kind-cluster.out))
	kind delete cluster
endif
	${RM} -r target
