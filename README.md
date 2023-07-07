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
 * Raw value in a separate file.
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
### GCP Workload identity

When running in GCP you can use the GCP Workload identity to authenticate to Vault. This requires that the [GCP Auth Method](https://www.vaultproject.io/docs/auth/gcp) is enabled in Vault and your service account has been given access to secrets. 
Check this blog post for more info : [Serverless Secrets with Google Cloud Run and Hashicorp Vault](https://bestseller.tech/posts/google-serverless-secrets/) 

To use, set the `gcpWorkloadID` flag to `true`.

---
<br/>

## Usage
In harpocrates you can specify which secrets to pull in 3 different ways.
### YAML file
yaml is a great options for readability and replication of configs. yaml options are: 

| Option          | Required | Value                                                        | default      |
|-----------------|----------|--------------------------------------------------------------|--------------|
| format          | no       | one of: env, json, secret                                    | env          |
| output          | yes      | /path/to/output/folder                                       | -            |
| owner           | no       | UID of the user e.g 0, can be set on "root" and secret level | current user |
| prefix          | no       | prefix, can be set on any level                              | -            |
| uppercase       | no       | will uppercase prefix and key                                | false        |
| append          | no       | appends secrets to a file                                    | true         |
| secrets         | yes      | an array of secret paths                                     | -            |
| gcpWorkloadID | no       | GCP workload identity, useful when running in GCP            | false        |

<br/>

Example yaml file at [examples/secret.yaml](examples/secret.yaml)

<br/>

run harpocrates with the `-f` flag to fetch secrets from your yaml spec.
```bash
harpocrates -f /path/to/file.yaml
```

<br/>

### Inline spec
You can specify the exact same options in inline json/yaml as in the yaml spec.
Mostly for programmatic use, as readability is way worse than the yaml spec.

```bash
harpocrates '{"format":"env","output":"/secrets","prefix":"PREFIX_","secrets":["secret/data/secret/dev",{"secret/data/foo":{"keys":["APIKEY"]}}]}'
```
Or if you prefer you can do it like this:
```bash
harpocrates '{
  "format": "env",
  "output": "/secrets",
  "prefix": "PREFIX_",
  "secrets": [
    "secret/data/secret/dev",
    {
      "secret/data/foo": {
        "keys": [
          "APIKEY"
        ]
      }
    }
  ]
}'
```

Or as yaml
```bash
harpocrates 'format: env
output: "/secrets"
prefix: PREFIX_
secrets:
  - secret/data/secret/dev
  - secret/data/foo:
      prefix: TEST_
      keys:
       - APIKEY
       - BAR:
           prefix: "BOTTOM_"
       - TOPSECRET:
           saveAsFile: true
  - secret/data/bar:
      format: json
      filename: something.json
      owner: 29'
```

<br/>

### CLI Parameters
The third option is to specify the options as parameters in the cli.

```bash
harpocrates --format "env" --secret "/secret/data/somesecret" --prefix "PREFIX_" --output "/secrets"
```
There is not the same granularity as in the json and yaml specs. e.g. prefix can only exist on the top level.

---
<br/>

## CLI and ENV Options

| Flag          | Env Var              | Values                                                                                                     |                       Default                       |
|---------------|----------------------|------------------------------------------------------------------------------------------------------------|:---------------------------------------------------:|
| vault-address | VAULT_ADDR           | https://vaulturl                                                                                           |                          -                          |
| auth-name     | AUTH_NAME            | Vault auth name, used at login                                                                             |                          -                          |
| role-name     | ROLE_NAME            | Vault role name, used at login                                                                             |                          -                          |
| token-path    | TOKEN_PATH           | /path/to/token, uses clustername and path to login and exchange a vault token which is used in vault_token | /var/run/secrets/kubernetes.io/serviceaccount/token |
| vault-token   | VAULT_TOKEN          | token as a string. If empty token_path will be queried                                                     |                          -                          |
| format        | -                    | env, json or secret                                                                                        |                         env                         |
| output        | -                    | /path/to/output                                                                                            |                  /tmp/secrets.env                   |
| owner         | -                    | UID of the user e.g 0                                                                                      |                    current user                     |
| prefix        | -                    | prefix keys, eg. K8S_                                                                                      |                          -                          |
| uppercase     | -                    | will uppercase prefix and key                                                                              |                        false                        |
| secret        | -                    | vault path /secretengine/data/some/secret                                                                  |                          -                          |
| append        | -                    | Appends secrets to a file                                                                                  |                        true                         |
| -             | HARPOCRATES_FILENAME | overwrites the default output filename                                                                     |                          -                          |
| gcpWorkloadID | GCP_WORKLOAD_ID      | set to true to enable GCP workload identity, useful when running in GCP                                    |                        false                        |
| -             | CONTINUOUS           | set to true to run harpocrates in a loop and fetch secrets every 1 minute, useful as a sidecar             |                        false                        |
| -             | INTERVAL             | set the interval in minutes for the continuous mode                                                        |                          1                          |
---
<br/>

## Kubernetes
When running `harpocrates` or `cloudrun` as an init container or sidecar you have to mount a volume to pass on the exported secrets to your main application.
Then you can either chose to source the env file or simply just read the json formatted file.
Harpocrates will startup and export the secrets in a matter of seconds. 

An example can be found at [examples/deployment.yaml](examples/deployment.yaml)

### Sidecar
To run harpocrates as a sidecar you have to set the `CONTINUOUS`  env var to true. Harpocrates will then run in a loop and fetch secrets every 1 minute. The shortest secret refresh interval is 1 minute and can be increased using the `INTERVAL` variable.

---
<br/>

## CircleCI Orb
Docs in the [orb folder](orb/README.md)


---
<br/>

## Contribution
Issues and pull requests are more than welcome.
