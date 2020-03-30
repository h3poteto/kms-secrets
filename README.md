[![CircleCI](https://circleci.com/gh/h3poteto/kms-secrets.svg?style=svg)](https://circleci.com/gh/h3poteto/kms-secrets)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/h3poteto/kms-secrets)](https://github.com/h3poteto/kms-secrets/releases)
[![Dependabot](https://img.shields.io/badge/Dependabot-enabled-blue.svg)](https://dependabot.com)

# KMS Secrets

KMS Secrets is custom controller for Kubernetes. This controller decrypts AWS KMS encrypted data and generate Secret resources. So you can apply KMS encrypted data definitions as it is.

## Overview
Sometimes you don't want to commit Secret resources to Git as is, because Secret has raw (base64 encoded) strings.
This request will often occur in GitOps workflow.

If you install KMS Secrets in your Kubernetes cluster, you can apply encrypted Secret resources as it is.
KMS Secrets will automatically decrypts these data and generate Secret resources, after you apply it.

#### Difference from SealedSecrets
This controller similar to [SealedSecrets](https://github.com/bitnami-labs/sealed-secrets). But KMS Secrets uses keys in AWS KMS instead of custom certificates to decrypt data.
This controller does not create public keys, certificates and private keys.
Therefore, you don't need manage public/private keys and certificates, and you can control decrypt permissions only with AWS IAM.

And KMS Secrets provides only custom controller, no CLI. So please use other tools to encrypt your data using AWS KMS, for example [kubesec](https://github.com/shyiko/kubesec), [yaml_vault](https://github.com/joker1007/yaml_vault) or [aws-cli](https://docs.aws.amazon.com/cli/latest/reference/kms/index.html).

## How to use it
Please define `KMSSecret` resource, like this:


```yaml
apiVersion: secret.h3poteto.dev/v1beta1
kind: KMSSecret
metadata:
  name: mysecret
  namespace: mynamespace
spec:
  encryptedData:
    API_KEY: AQICAHh2iCEGE2e6vdC+w6dQ4hRIyahEPE...
    PASSWORD: AQICAHh2iCEGE2e6vdC+w6dQ4hRIyahEPE...
  region: us-east-1
  template:
    metadata:
      labels:
        "h3poteto.dev/custom-labels": my-label
      annotations:
        "h3poteto.dev/annotations": my-annotation
```

In this time, please provide KMS encrypted strings in `encryptedData`. You can get KMS encrypted strings using [kubesec](https://github.com/shyiko/kubesec), [yaml_vault](https://github.com/joker1007/yaml_vault) or [aws-cli](https://docs.aws.amazon.com/cli/latest/reference/kms/index.html).

Here is an example using aws-cli.

```
$ aws kms encrypt --key-id 1asdf3-rsdf... --plaintext "apikey" --query CiphertextBlob --output text
AQICAHh2iCEGE2e6vdC+w6dQ4hRIyahEPE...
```

Please provide raw text, you don't need to provide base64 encoded strings. Because aws command outputs base64 encoded strings through KMS decrypt.


And if you provide `spec.template.metadata`, `labels` and `annotations` are applied to generated Secret.


After you apply `KMSSecret`, the custom controller will generate a Secret which has same name and namespace as `KMSSecret`, like this:

```yaml
apiVersion: v1
data:
  API_KEY: YXBpa2V5
  PASSWORD: cGFzc3dvcmQ=
kind: Secret
metadata:
  annotations:
    h3poteto.dev/annotations: my-annotation
  creationTimestamp: "2020-03-18T07:27:06Z"
  labels:
    h3poteto.dev/custom-labels: my-label
  name: mysecret
  namespace: mynamespace
  ownerReferences:
  - apiVersion: secret.h3poteto.dev/v1beta1
    blockOwnerDeletion: true
    controller: true
    kind: KMSSecret
    name: mysecret
    uid: deac9220-68e9-11ea-8182-0658b210029a
  resourceVersion: "10673189"
  selfLink: /api/v1/namespaces/mynamespace/secrets/mysecret
  uid: dec32aea-68e9-11ea-ae9a-0a561536d7cc
type: Opaque
```



## How to install
### Helm

You can install KMS Secrets using helm:

```
$ helm repo add h3poteto-stable https://h3poteto.github.io/charts/stable
$ helm install h3poteto-stable/kms-secrets --name kms-secrets
```

And please refer configuration on [chart repository](https://github.com/h3poteto/charts/tree/master/stable/kms-secrets).

### Kustomize
Kustomize template is in [config](/config/default).
You can use `kucectl` has native support for [kustomize](https://kustomize.io/).

```
$ git clone https://github.com/h3poteto/kms-secrets.git
$ cd kms-secrets
$ kubectl apply -k config/default
```

Probably you have to customize ServiceAccount's annotations to enable [IAM Role for Service Account](https://aws.amazon.com/blogs/opensource/introducing-fine-grained-iam-roles-service-accounts/), because this controller requests access to AWS KMS. Please fill it with your IAM Role.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: manager
  namespace: system
  annotations:
    eks.amazonaws.com/role-arn:  arn:aws:iam::123456789:role/your-iam-role
```

### IAM Policy
Your IAM Role which is assigned KMS Secrets controller, requires this policy.

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "Allow use of the key",
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt",
        "kms:DescribeKey"
      ],
      "Resource": "*"
    }
  ]
}
```

## License
The package is available as open source under the terms of the [MIT License](https://opensource.org/licenses/MIT).
