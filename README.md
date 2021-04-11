# image-clone-controller

`kubectl get po -o=jsonpath='{range .items[*]}{.spec.containers[*].image}{"\n"}{end}'`