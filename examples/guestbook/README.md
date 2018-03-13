# Description

Example from the Stateless Applications deployment into Kubernetes - Guestbook.

https://kubernetes.io/docs/tutorials/stateless-application/guestbook/

```bash
aptomictl login -u admin -p admin
aptomictl policy apply --wait -f ~/.aptomi/examples/guestbook/acl.yaml

aptomictl login -u john -p john
aptomictl policy apply --wait -f ~/.aptomi/examples/guestbook/guestbook.yaml

aptomictl login -u alice -p alice
aptomictl policy apply --wait -f ~/.aptomi/examples/guestbook/alice-guestbook.yaml

aptomictl login -u bob -p bob
aptomictl policy apply --wait -f ~/.aptomi/examples/guestbook/bob-guestbook.yaml
```
