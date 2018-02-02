# Compact Mode
* Aptomi will be installed on a local machine 

# Installation

## Binaries on local machine
Aptomi has an installer script that will automatically install & configure the latest versions of Aptomi server & client. You could also download that script, inspect and execute it. It's well documented, so you can read through it and understand what it is doing before you run it.
```bash
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_install.sh | bash && aptomi server
```
Once Aptomi server is started, it will serve API and UI on default port 27866.

## Docker container on a local machine
Alternatively, you can run Aptomi server in a Docker container: 
```bash
docker run -it --rm -p 27866:27866 aptomi/aptomi-test-install:xenial sh -c 'curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_install.sh | bash && aptomi server'
```

And install client locally:
```

```

# Accessing UI
Open UI at [http://localhost:27866/](http://localhost:27866/) and log in as **'admin/admin'**. It's a pre-configured Aptomi domain admin user with full access rights. Once you get going and set up more admin users, you can disable this account or change password later on.

At this point, most UI screens will be empty. This is expected, as Aptomi has no applications imported yet.

Now you are ready to move on to importing the examples and start deploying apps.
