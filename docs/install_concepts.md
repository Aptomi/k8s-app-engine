# Aptomi Install / Concepts Mode
* Aptomi will be installed on a local machine and pre-populated with an example
* The part responsible for application deployment **will be disabled**. It means you can explore Aptomi without configuring a k8s cluster, but you won't be able to deploy any apps either

# Installation

## Option #1: local binaries
Install Aptomi locally, pre-populated with an example:
```bash
curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_install.sh | bash /dev/stdin --with-example && aptomi server
```

## Option #2: docker container
Alternatively, you can run Aptomi server pre-populated with an example in a Docker container: 
```bash
docker run -it --rm -p 27866:27866 aptomi/aptomi-test-install:xenial sh -c 'curl https://raw.githubusercontent.com/Aptomi/aptomi/master/scripts/aptomi_install.sh | bash /dev/stdin --with-example && aptomi server'
```

# Accessing UI
Open UI at [http://localhost:27866/](http://localhost:27866/) and log in as **'admin/admin'**.

Example is already loaded into Aptomi, so you don't need to load it separately. 
