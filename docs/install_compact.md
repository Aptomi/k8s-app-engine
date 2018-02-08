# Aptomi Install / Compact Mode
* Aptomi will be installed on a local machine 

# Installation

## Option #1: Aptomi server in Docker container & client - local binary
You can run Aptomi **server** in a Docker container: 
```bash
docker run -it --rm -p 27866:27866 aptomi/aptomi-test-install:xenial sh -c 'curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_install.sh | bash && aptomi server'
```

And install **client** locally:
```
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_install.sh | bash /dev/stdin --client-only
```

## Option #2: Aptomi server & client - local binaries
Alternatively, you can install Aptomi server and client on a local machine (you can read through the script if you have doubts):
```bash
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_install.sh | bash && aptomi server
```

It will install:
* Aptomi binaries in `/usr/local/bin/`
* Aptomi server config in `/etc/aptomi/` and use `/var/lib/aptomi` as persistent data store
* Aptomi client config and examples in `~/.aptomi/`

# Accessing UI
Once Aptomi server is started, it will serve API and UI on default port 27866.

Open UI at [http://localhost:27866/](http://localhost:27866/) and log in as **'admin/admin'**. It's a pre-configured Aptomi domain admin user with full access rights. Once you get going and set up more admin users, you can disable this account or change password later on.

At this point, most UI screens will be empty. This is expected, as Aptomi has no applications imported yet.

Now you are ready to move on to point Aptomi to your k8s cluster(s) and start deploying your apps.

# Useful Commands

## Cleaning up local installation
To delete all Aptomi binaries installed locally as well as Aptomi data, run:
```
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_uninstall_and_clean.sh | bash
```
