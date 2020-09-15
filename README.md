# Harpocrates
> Harpocrates was the god of silence, secrets and confidentiality

<br/>

![CircleCI](https://img.shields.io/circleci/build/github/BESTSELLER/harpocrates/master)
![GitHub repo size](https://img.shields.io/github/repo-size/BESTSELLER/harpocrates)
![GitHub All Releases](https://img.shields.io/github/downloads/BESTSELLER/harpocrates/total)
![GitHub](https://img.shields.io/github/license/BESTSELLER/harpocrates)

<br/>

Harpocrates is a small application that can be used to pull secrets from [HashiCorp Vault](https://www.vaultproject.io/).
It can output the secrets in different formats:
 * JSON, which is simple key-values.
 * `source` ready env file e.g.
 ```bash
 export KEY=value
 export FOO=bar
 ```
 * Raw key values.
 * Raw value in separate file.

<br/><br/>

Harpocrates is designed such it can be used as an init- or sidecar container in [Kubernetes](https://kubernetes.io/). 
In this scenario it uses the ServiceAccount token in `/var/run/secrets/kubernetes.io/serviceaccount/token` and exchanges this for a Vault token by posting it to `/auth/kubernetes/login`.

This requires that the [Kubernetes Auth Method](https://www.vaultproject.io/docs/auth/kubernetes) is enabled in Vault.

---
<br/>

## Authentication
The easiest way to authenticate is to use your Vault token:
```bash
harpocrates --vault-token "sometoken"
```
This can also be specified as the environment var `VAULT_TOKEN`

---
<br/>

## Usage
In harpocrates you can specify which secrets to pull in 3 different ways.
### YAML file
yaml is a great options for readability and replication of configs. yaml options are: 

| Option  | Required | Value                           | default |
| ------- | -------- | ------------------------------- | ------- |
| format  | no       | one of: env, json, secret       | env     |
| output  | yes      | /path/to/output/folder          | -       |
| prefix  | no       | prefix, can be set on any level | -       |
| secrets | yes      | an array of secret paths        | -       |

<br/>

Example yaml file:
```yaml
format: env
output: "/secrets"
prefix: PREFIX_
secrets:
  - secret/data/secret/dev # Will pull all key-values from the secret path.
  - secret/data/foo:
      prefix: TEST_ # overwrites the toplevel prefix
      keys:
       - APIKEY # fetches only this specific key and value from `secret/data/foo`
       - BAR:
           prefix: "BOTTOM_" # overwrites both secret and toplevel prefix.
       - TOPSECRET:
           saveToFile: true # saves ONLY the raw value to a file, which is named as the key.
```

<br/>

run harpocrates with the `-f` flag to fetch secrets from your yaml spec.
```bash
harpocrates -f /path/to/file.yaml
```

<br/>

### Inline JSON
You can specify the exact same options in inline json as in the yaml spec.
Mostly for programmatic use, as readability is way worse than the yaml spec.

```bash
harpocrates '{"format":"env","output":"/secrets","prefix":"PREFIX_","secrets":["secret/data/secret/dev",{"secret/data/foo":{"keys":["APIKEY"]}}]}'
```

<br/>

### CLI Parameters
Third option is to specify the options as parameters in the cli.

```bash
harpocrates --format "env" --secret "/secret/data/somesecret" --prefix "PREFIX_" --output "/secrets"
```
There are not the same granularity as in the json and yaml specs. e.g. prefix can only exist on the top level.

---
<br/>

## CLI and ENV Options

| Flag          | Env Var              | Values                                                                                                     |                       Default                       |
| ------------- | -------------------- | ---------------------------------------------------------------------------------------------------------- | :-------------------------------------------------: |
| vault-address | VAULT_ADDR           | https://vaulturl                                                                                           |                          -                          |
| cluster-name  | CLUSTER_NAME         | string                                                                                                     |                          -                          |
| token-path    | TOKEN_PATH           | /path/to/token, uses clustername and path to login and exchange a vault token which is used in vault_token | /var/run/secrets/kubernetes.io/serviceaccount/token |
| vault-token   | VAULT_TOKEN          | token as a string. If empty token_path will be queried                                                     |                          -                          |
| format        | -                    | env, json or secret                                                                                        |                         env                         |
| output        | -                    | /path/to/output                                                                                            |                  /tmp/secrets.env                   |
| prefix        | -                    | prefix keys, eg. K8S_                                                                                      |                          -                          |
| secret        | -                    | vault path /secretengine/data/some/secret                                                                  |                          -                          |
| -             | HARPOCRATES_FILENAME | overwrites the default output filename                                                                     |                          -                          |


---
<br/>

## Kubernetes
When running `harpocrates` as an init container you have to mount a volume to pass on the exported secrets to your main application.
Then you can either chose to source the env file or simply just read the json formatted file.
Harpocrates will startup and export the secrets in a matter of seconds. 

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    cluster-autoscaler.kubernetes.io/safe-to-evict: "true" # isset as the autoscaler will stop downscaling on pods with local volumes
  labels:
    app: test-app
  name: test-app
  namespace: default
spec:
  selector:
    matchLabels:
      app: test-app
  template:
    metadata:
      annotations:
        cluster-autoscaler.kubernetes.io/safe-to-evict: "true" # isset as the autoscaler will stop downscaling on pods with local volumes
      labels:
        app: test-app
    spec:
      containers:
        - image: mytestapp
          name: test-app
          volumeMounts:
            - mountPath: /secrets # mount our volume with secrets
              name: secrets
      initContainers: # the fun part
        - args: # inline json is the easiest when running as init
            - '{"format":"env","output":"/secrets","prefix":"PREFIX_","secrets":["secret/data/secret/dev",{"secret/data/foo":{"keys":["APIKEY"]}}]}'
        env:
          - name: VAULT_ADDR # which vault server it should connect to
            value: https://vault.example.com
          - name: CLUSTER_NAME # the kubernetes auth method
            value: testcluster
        image: harbor.bestsellerit.com/library/harpocrates
        name: secret-dumper
        volumeMounts: # mount the volume to export secrets
          - mountPath: /secrets
            name: secrets
      volumes: # volume spec
        - name: secrets
          emptyDir: {}
```

---
<br/>

## CircleCI Orb
We have created a CircleCI Orb which utilizes `harpocrates`for secret injection in both our kubernetes deployments and circleci jobs.
The orb some parts that is tailored to fit our own usecases, but still useable for others as well.

Find the Orb  and docs [here]()

<br/>

**Kubernetes deployments**

We use Kustomize to append the init container and volume section to kubernetes specs. This is dependent on the deployment and container name matches the specified values in our harpocrates YAML spec. It will only append the secret volume to the container that matches the name in the harpocrates YAML spec.

<br/>

***note***

We have to set the following annotation, in order for the kubernetes autoscaler to be able to scale down again.
```
annotations:
    "cluster-autoscaler.kubernetes.io/safe-to-evict": "true"
```
https://issuetracker.google.com/issues/148295270


---
<br/>

## Contribution
Issues and pull requests are more than welcome.
