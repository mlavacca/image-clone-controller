# Image-clone-controller

[![asciicast](https://asciinema.org/a/SrgRSAmIx2JOUs14GazEoiy2n.svg)](https://asciinema.org/a/SrgRSAmIx2JOUs14GazEoiy2n)

```bash
kubectl get deploy -A -o=jsonpath='{range .items[*]}{.metadata.namespace}{"\t\t"}{.spec.template.spec.containers[*].image}{"\n"}{end}'
```
```bash
kubectl get daemonset -A -o=jsonpath='{range .items[*]}{.metadata.namespace}{"\t\t"}{.spec.template.spec.containers[*].image}{"\n"}{end}'
```
