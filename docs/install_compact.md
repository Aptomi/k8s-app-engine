# Compact Mode
* Aptomi binaries will be installed on a local machine
* Apps can be deployed via Aptomi to any local or remote k8s (minikube, docker for mac, GKE, etc)

# Installation
Aptomi has an installer script that will automatically get the latest version of Aptomi and install it locally. You could also download that script, inspect and execute it. It's well documented, so you can read through it and understand what it is doing before you run it.
```bash
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_install.sh | bash
```

# Starting Aptomi server
Start Aptomi server, it will serve API and UI on default port 27866:
```bash
aptomi server
```

Open UI at [http://localhost:27866/](http://localhost:27866/) and log in as **'admin/admin'**. It's a pre-configured Aptomi domain admin user with full access rights. Once you get going and set up more admin users, you can disable this account or change password later on.

At this point, most UI screens will be empty. This is expected, as Aptomi has no applications imported yet.

Now you are ready to move on to importing the examples and start deploying apps.
