### Installation
The best way to install Aptomi is to download its latest release, which contains compiled server and client binaries for various platforms:
- Aptomi Server is an all-in-one binary with embedded DB store, which serves API requests, runs UI, as well as does deployment and continuous state enforcement
- Aptomi Client is a client for talking to Aptomi Server. It allows end-users of Aptomi to feed YAML files into Aptomi Server over REST API

You can run those binaries locally.

Additionally you can download binary using `go get -u gopkg.in/Aptomi/aptomi.v0/cmd/aptomictl` command.
Make sure that your `$GOPATH/bin` is in the `$PATH` to use it.
You can rerun this command to update your client.

And finally it's possible just to use dockerized client in a following way:

```bash
# just run docker run directly:
docker run -it --rm -v "$HOME/.aptomi/":"/root/.aptomi" aptomi/aptomictl:0 policy show

# or add alias for it:
alias aptomictl='docker run -it --rm -v "$HOME/.aptomi/":"/root/.aptomi" aptomi/aptomictl:0'

# to update client you'll need to run:
docker pull aptomi/aptomictl:0
```

#### Configuring LDAP
Aptomi needs to be configured with user data source in order to enable UI login and make policy decisions based on users' labels/properties. It's recommended to
start with LDAP, which is also required by Aptomi examples and smoke tests.

1. LDAP Server with sample users is provided in a docker container. To download and start the published LDAP server image, run:
    ```
    ./tools/demo-ldap.sh
    ```
2. Even though it's not required, you may want to download and install [Apache Directory Studio](http://directory.apache.org/studio/) to familiarize yourself with the user data in provided in sample LDAP server. Once installed,
follow these [step-by-step instructions](http://directory.apache.org/apacheds/basic-ug/1.4.2-changing-admin-password.html) to connect to LDAP and browse it. Use default credentials given in the manual.

1. Download the latest release of Aptomi from [releases](https://github.com/Aptomi/aptomi/releases).
    It comes with server and client binaries as well as examples directory and needed tools. Unpack it into some directory:
    ```
    export aptomi_version=X.Y.Z
    export aptomi_os=darwin # or linux
    export aptomi_arch=amd64 # or 386
    export aptomi_name=aptomi_${aptomi_version}_${aptomi_os}_${aptomi_arch}

    wget https://github.com/Aptomi/aptomi/releases/download/v${aptomi_version}/${aptomi_name}.tar.gz
    tar xzf ${aptomi_name}.tar.gz
    cd ${aptomi_name}
    ```

1. Create config for Aptomi server and start it. It will serve API and UI :
    ```
    mkdir /var/lib/aptomi
    sudo cp examples/config/server.yaml /var/lib/aptomi/config.yaml
    aptomi server
    ```

1. Create config for Aptomi client and make sure it can connect to the server:
    ```
    mkdir ~/.aptomi
    cp examples/config/client.yaml ~/.aptomi/config.yaml
    aptomictl -u Sam policy show
    ```
    You should be able to see:
    ```
    &{{policy} {1 2017-11-19 00:00:05.613151 -0800 PST aptomi} map[]}
    ```
