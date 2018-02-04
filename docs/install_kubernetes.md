# Aptomi Install / Kubernetes Mode
* You must have k8s cluster ready to go
* Aptomi server will be installed in k8s via Helm chart
* Aptomi client will be installed locally

# Installing & Configuring Helm client
Our recommendation is to use Helm client 2.6.2 (as 2.8 doesn't  work too well):

```
curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash /dev/stdin -v v2.6.2
helm init --client-only
helm repo add aptomi http://aptomi.io/charts
helm repo update
```

# Installing & Configuring Aptomi
Figure out which context you want to install Aptomi server to (use `kubectl config get-contexts`). Then replace `[CONTEXT_NAME]` with the actual name and run:
```
helm install --kube-context [CONTEXT_NAME] --name aptomi --namespace aptomi aptomi/aptomi --set users.admin.enabled=true,users.example.enabled=true
```

Once Aptomi server is deployed, it will tell you the port it listens on: 
```
==> v1/Service
NAME           CLUSTER-IP     EXTERNAL-IP  PORT(S)          AGE
aptomi-aptomi  10.96.102.113  1.2.3.4      27866:56789/TCP  0s
```

Install Aptomi client locally:
```
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_install.sh | bash /dev/stdin --client-only
```

Configure Aptomi client to use deployed Aptomi server and test it:
```
vi ~/.aptomi/config.yaml

    api:
      host: 1.2.3.4  <- replace
      port: 56789    <- replace
      
aptomictl version
aptomictl policy show
```

# Accessing UI
Open UI at `http://APTOMI_SERVER_IP:PORT/` and log in as **'admin/admin'**. It's a pre-configured Aptomi domain admin user with full access rights. Once you get going and set up more admin users, you can disable this account or change password later on.

At this point, most UI screens will be empty. This is expected, as Aptomi has no applications imported yet.

Now you are ready to move on to point Aptomi to your k8s cluster(s) and start deploying your apps.

# Important note before you close this page

You will find in subsequent instructions how to point Aptomi to your k8s cluster(s).

If Aptomi server itself is running inside k8s cluster, it may not be able to communicate to the same k8s cluster via external `ip:port` from the inside.  

So, when generating cluster YAMLs and importing them into Aptomi, you will need to tell it to use a **local** cluster instead pointing to a specific kubectl context. Luckily, there is a corresponding CLI flag `-l`, so every time you see `aptomictl gen cluster` in the instructions:
* don't forget to use `aptomictl gen cluster -l ...`
* instead of `aptomictl gen cluster -c [CONTEXT_NAME]. ..`

# Useful Commands

## Upgrading Aptomi server from latest release -> master
```
helm upgrade aptomi aptomi/aptomi --reuse-values --set image.tag=master
```

## Restarting Aptomi server
```
helm upgrade aptomi aptomi/aptomi --reuse-values --recreate-pods
```
 