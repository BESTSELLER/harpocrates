# Harpocrates

This will be the home of the master of all secrets.


When using a ServiceAccount in Kubernetes, the jwt token can be retrieved by reading the file `/var/run/secrets/kubernetes.io/serviceaccount/token`

And then it can be exchanged to a Vault token by posting it to `/auth/kubernetes/login`

Example of a secret file:
```yaml
format: json
output: path/to/dir/to/save/secret/to
secrets:
  - path/to/secret1
  - path/to/secret2:
    - key1:
        saveAsFile: true
    - key2
```
At the moment it takes a json file as input, you can convert your secret to json by doing:
`yq read secret.yml -j`

Orb should read kustomize yaml from Vault


## Deployment.yml
To use this, you can add harpocrates as an initContainers like so:
```yaml
initContainers:
  - name: secret-dumper
    image: harbor.bestsellerit.com/library/harpocrates:68
    args:
      - '{"format":"env","output":"/secrets","prefix":"alfeios_","secrets":["ES/data/alfeios/prod"]}'
    volumeMounts:
      - name: secrets
        mountPath: /secrets
    env:
      - name: VAULT_ADDRESS
        value: $VAULT_ADDR
      - name: CLUSTER_NAME
        value: es03-prod
volumes:
  - name: secrets
    emptyDir: {}
```

CircleCI steps:
```yaml
- secret-injector:
    app-name: alfeios
    file: deployment.yml
    secretFile: alfeios-secrets.yml
- secret-injector:
    app-name: alfeios-db
    file: deployment.yml
    secretFile: alfeios-db-secrets.yml
```


## TO-DO

* Support files ?


## NOTES
We have to set the following annotation, in order for the autoscaler to be able to scale down again.
```
annotations:
    "cluster-autoscaler.kubernetes.io/safe-to-evict": "true"
```
https://issuetracker.google.com/issues/148295270