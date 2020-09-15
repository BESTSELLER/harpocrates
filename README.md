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
    ```json
    {
      "KEY": "value",
      "FOO": "bar"
    }
    ```
 * `source` ready env file e.g.
    ```bash
    export KEY=value
    export FOO=bar
    ```
 * Raw key values.
    ```bash
    KEY=value
    FOO=bar
    ```
 * Raw value in separate file.
    ```bash
    value
    ```

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

Example yaml file at [examples/secret.yaml](examples/secret.yaml)

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

An example can be found at [examples/deployment.yaml](examples/deployment.yaml)

---
<br/>

## CircleCI Orb
Docs in the [orb folder](orb/README.md)


---
<br/>

## Contribution
Issues and pull requests are more than welcome.
