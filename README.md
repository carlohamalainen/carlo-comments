# carlo-comments
Commenting system for Carlo's personal blog

https://github.com/0xdod/go-realworld

https://www.digitalocean.com/community/tutorials/how-to-secure-your-site-in-kubernetes-with-cert-manager-traefik-and-let-s-encrypt


# Install notes

Ensure all secrets and configs set:

```
./set-k8-secrets.sh
```

```shell
docker build -t carlo-comments:v1 .
docker tag carlo-comments:v1 registry.digitalocean.com/carlo-containers/carlo-comments:v1
docker push registry.digitalocean.com/carlo-containers/carlo-comments:v1

```shell
$ docker image ls
REPOSITORY                                                  TAG       IMAGE ID       CREATED         SIZE
carlo-comments                                              v1        e24335217b8b   2 minutes ago   26.5MB
registry.digitalocean.com/carlo-containers/carlo-comments   v1        e24335217b8b   2 minutes ago   26.5MB
```



```shell
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update ingress-nginx

NGINX_CHART_VERSION="4.1.3"

helm install ingress-nginx ingress-nginx/ingress-nginx --version "$NGINX_CHART_VERSION" \
  --namespace ingress-nginx \
  --create-namespace \
  -f "nginx-values-v${NGINX_CHART_VERSION}.yaml"
```

quirk with DigitalOcean load balancers, getting EOF on challenge

https://linuxblog.xyz/posts/installing-nginx-ingress-digitalocean/

```shell
NGINX_CHART_VERSION="4.1.3"

helm install ingress-nginx ingress-nginx/ingress-nginx --version "$NGINX_CHART_VERSION" \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=LoadBalancer \
  --set controller.service.externalTrafficPolicy=Local \
  --set controller.config.use-proxy-protocol="true" \
  --set-string controller.service.annotations."service\.beta\.kubernetes\.io/do-loadbalancer-enable-proxy-protocol"="true" \
  --set-string controller.service.annotations."service\.beta\.kubernetes\.io/do-loadbalancer-hostname"="workaround.example.org" \
  -f "nginx-values-v${NGINX_CHART_VERSION}.yaml"

```












Check for load balancer external IP:

```shell
kubectl get svc -n ingress-nginx
```

```
helm repo add jetstack https://charts.jetstack.io
helm repo update jetstack

helm install \
  cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.14.5 \
  --set installCRDs=true
```


```shell
kubectl create ns backend
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f host.yaml
kubectl apply -f issuer-letsencrypt-nginx.yaml

```

Check general status (note ``0/2`` ready, indicating failure):

```
$ kubectl get deployments --all-namespaces
NAMESPACE       NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
backend         carlo-comments             0/2     2            0           13m
cert-manager    cert-manager               1/1     1            1           9m44s
cert-manager    cert-manager-cainjector    1/1     1            1           9m44s
cert-manager    cert-manager-webhook       1/1     1            1           9m44s
ingress-nginx   ingress-nginx-controller   2/2     2            2           21m
kube-system     coredns                    2/2     2            2           26m
kube-system     hubble-relay               1/1     1            1           28m
kube-system     hubble-ui                  1/1     1            1           27m
```