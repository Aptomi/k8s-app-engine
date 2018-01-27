# Compact Mode
* Aptomi binaries will be installed on a local machine
* Apps can be deployed via Aptomi to any local or remote k8s (minikube, docker for mac, GKE, etc)
* Prerequisites - docker

# Installation
Aptomi has an installer script that will automatically get the latest version of Aptomi and install it locally. You could also download that script and execute it locally. It's well documented, so you can read through it and understand what it is doing before you run it.
```bash
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_install.sh | bash
```

# Start services
You must start a supplied LDAP container locally, so you can run through the examples provided with Aptomi:
```bash
docker run --name aptomi-ldap-demo -d -p 10389:10389 aptomi/ldap-demo:latest
```

Ensure that status of LDAP container is "Up":
```bash
docker ps -a
```

Start Aptomi server, it will server API and UI on default port 27866:
```bash
aptomi server
```

Open UI at [http://localhost:27866/](http://localhost:27866/) and log in as 'sam:sam'. It's a built-in Aptomi administrator. Once you set up other users, you can disable this user account or change password later on.

At this policy UI screens will be mostly empty, as Aptomi has no objects imported. At this point you are ready to move on to importing the examples.

# Common Issues

## Status of LDAP container is "Exited"
If the status of LDAP container is "Exited", then you likely have an issue with Docker itself not properly working on your machine.
You can still look at the logs of LDAP container, but you will likely find a one-liner error there:
```bash
docker logs aptomi-ldap-demo
```

## Unable to login into UI (check username/password)
Likely there is a connection issue to LDAP. Check Aptomi server logs for:
```
ERRO[0000] Error while serving request: LDAP Result Code 200 "Network Error": dial tcp [::1]:10389: getsockopt: connection refused
```