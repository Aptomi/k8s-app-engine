# Compact Mode
* Aptomi binaries will be installed on a local machine
* Apps can be deployed via Aptomi to any local or remote k8s (minikube, docker for mac, GKE, etc)

# Installation
Aptomi has an installer script that will automatically get the latest version of Aptomi and install it locally. You could also download that script and execute it locally. It's well documented, so you can read through it and understand what it is doing before you run it.
```bash
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_install.sh | bash
```

# Starting Aptomi server
Start Aptomi server, it will serve API and UI on default port 27866:
```bash
aptomi server
```

Open UI at [http://localhost:27866/](http://localhost:27866/) and log in as **'admin/admin'**. It's a pre-configured Aptomi domain admin. Once you get going and set up other users, you can disable this account or change password later on. At this point,
most UI screens will be empty. This is expected, as Aptomi has no objects imported yet.

# Setting up LDAP
Aptomi application examples use LDAP as source of user data. In order to run examples you **must** start Aptomi LDAP Demo server in a Docker container and configure LDAP data source in Aptomi:
* Stop Aptomi server (CTRL+C works)
* Start Aptomi LDAP Demo server and ensure the status of its container is "Up"
  * ```bash
    docker run --name aptomi-ldap-demo -d -p 10389:10389 aptomi/ldap-demo:latest
    ```
  * ```bash
    docker ps -a
    ```
* Change Aptomi configuration to enable LDAP and start Aptomi Server again
  * ```bash
    sudo sed -i.bak -e 's/ldap-disabled/ldap/g' /etc/aptomi/config.yaml
    ```
  * ```bash
    aptomi server
    ```

Open UI at [http://localhost:27866/](http://localhost:27866/), log out, and log in as **'sam/sam'**. It's a user from Aptomi LDAP Demo server.

If the log in is successful, now you are ready to move on to importing the examples and deploying apps.

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
