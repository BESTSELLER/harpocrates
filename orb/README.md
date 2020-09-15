## CircleCI Orb
We have created a CircleCI Orb which utilizes `harpocrates`for secret injection in both our kubernetes deployments and circleci jobs.
The orb some parts that is tailored to fit our own usecases, but still useable for others as well.

Find the Orb  and docs [here](https://circleci.com/orbs/registry/orb/bestsellerit/secret-injector)

<br/>

**Kubernetes deployments**

We use Kustomize to append the init container and volume section to kubernetes specs. This is dependent on the deployment and container name matches the specified values in our harpocrates YAML spec. It will only append the secret volume to the container that matches the name in the harpocrates YAML spec.

<br/>

***Note***

We have to set the following annotation, in order for the kubernetes autoscaler to be able to scale down again.
```
annotations:
    "cluster-autoscaler.kubernetes.io/safe-to-evict": "true"
```
https://issuetracker.google.com/issues/148295270


---
<br/>