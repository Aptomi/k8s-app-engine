# Aptomi Install / Compact Mode
* Aptomi will be installed on a local machine 

# Installation

## Option #1: Aptomi server in Docker container & client - local binary
You can run the Aptomi **server** in a Docker container: 
```bash
docker run -it --rm -p 27866:27866 aptomi/aptomi-test-install:xenial sh -c 'curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_install.sh | bash && aptomi server'
```

And install the **client** locally:
```
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_install.sh | bash /dev/stdin --client-only
```

## Option #2: Aptomi server & client - local binaries
Alternatively, you can install the Aptomi server and client on a local machine (we encourage you to read through the script if you have doubts!):
```bash
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_install.sh | bash && aptomi server
```

This will install:
* Aptomi binaries in `/usr/local/bin/`
* Aptomi server config in `/etc/aptomi/` and use `/var/lib/aptomi` as persistent data store
* Aptomi client config and examples in `~/.aptomi/`

# Accessing the UI
Once the Aptomi server is started, it will serve the API and UI on default port 27866.

Open the UI at [http://localhost:27866/](http://localhost:27866/) and log in as **'admin/admin'**. This is a pre-configured Aptomi domain admin user with full access rights. Once you get going and set up more admin users, you can disable this account or change its password later on.

At this point, most of the UI screens will be empty. This is expected, as Aptomi has no applications imported yet.

# Next Steps
You are now ready to point Aptomi to your k8s cluster(s) and start deploying your apps!

Depending on your k8s cluster, you can:

* Configure Aptomi to use [an existing k8s cluster](k8s_own.md)
* Configure Aptomi to use [GKE](k8s_gke.md)
* Configure Aptomi to use [Minikube](k8s_minikube.md), or
* Configure Aptomi to use [Docker For Mac](k8s_docker_for_mac.md) 

# Useful Commands

## Cleaning up local installation
To delete all Aptomi binaries installed locally on your machine, as well as all Aptomi data, run:
```
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_uninstall_and_clean.sh | bash
```
