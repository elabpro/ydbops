## ydb-ops

#### Disclaimer: heavy work in progress, expect changes until probably March 20th

If you have noticed something in code that you didn't like, then please tell jorres@, but it's probably in refactoring plans anyway.

#### How to build:

No special actions, just do from the repo root:

```
go build
```

#### Currently unimplemented, but will be in the nearest future:

1. The tests (`ginkgo test -vvv ./tests`) currently use real `time`, which means they run for a while (10-15 seconds). I will create a fake clock for tests later.
2. Yandex IAM authorization with SA account key file is currently unsupported.
   However, you can always issue the token yourself and put it inside the `YDB_TOKEN`
   variable: `export YDB_TOKEN=$(ycp --profile <profile> iam create-token)`
3. `restart tenant k8s` is not implemented yet, but soon.
4. `restart tenant` command in general is LEAST tested. Expect bugs.
5. For prettiness: The module './pkg/rolling/restarters' will be probably split into
   multiple modules, one module per restarter.

#### How to use:

Please browse the `ydbops --help` first. Then read along for examples (substitute your own values, of course).

##### Restart baremetal storage hosts

```
ydbops restart storage baremetal \
  --endpoint grpc://<cluster-fqdn> \
  --ssh-args=pssh,-A,-J,<bastion-fqdn>,--ycp-profile,prod,--no-yubikey \
  --verbose --hosts <node1-fqdn>,<node2-fqdn>,<node3-fqdn>
```

##### Run hello-world on remote hosts

```
ydbops restart run \
  --endpoint grpc://<cluster-fqdn> \
  --availability-mode strong --verbose --hosts 7,8 \
  --payload ./tests/payloads/payload-echo-helloworld.sh
```

##### Restart hosts using a custom payload

```
ydbops restart run \
  --endpoint grpc://<cluster-fqdn> \
  --availability-mode strong --verbose --hosts 5,6 \
  --payload ./tests/payloads/payload-restart-ydbd.sh
```

##### Restart storage in k8s

An example of authenticating with static credentials:

```
export YDB_PASSWORD=password_123
ydbops restart storage k8s \
  --endpoint grpc://<cluster-fqdn> \
  --availability-mode strong --verbose --hosts 7,8 \
  --user jorres --kubeconfig ~/.kube/config
```

## Tests

TODO jorres@ describe the current expressive power of tests and how to write them.

Please don't look into 'black-magic.go' for now. I promise it's going to become prettier.