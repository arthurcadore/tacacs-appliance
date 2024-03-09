# TACACS+ Server Appliance with docker

### This repository implements the [TACACS+](https://hub.docker.com/r/lfkeitel/tacacs_plus) images avaiable on DockerHub to host an mantain an AAA server.

Prerequisites
Before you begin, ensure you have the following packages installed on your system:

- Git version 2.34.1
- Docker version 24.0.6, build ed223bc
- Docker Compose version v2.21.0

---
### Getting Started:

First, copy the line below and paste on your prompt to clone the repository:

```
git clone https://github.com/arthurcadore/capacita-tacacs
```
If you don't have installed the package Git yet, do it before try to clone the respository!

Navigate to the project directory:

```
cd ./capacita-tacacs
```

If you don't have Docker (and Docker-compose) installed on your system yet, it can be installed by run the following commands (Script for Ubuntu 22.04): 

```
./installDocker.sh
```

**If you had to install docker, please remember to reboot you machine to grant user privileges for docker application.** 

In sequence, configure the environment files for the application container, you can do this by edditing the following files: 

#### config/tac_plus.cfg -> Change the TACACS+ Server parameters for host access:
```
    host = world {
        address = 0.0.0.0/0
        enable = clear enable
        key = tac_plus_key
    }
```

#### config/tac_plus.cfg -> Change the TACACS+ Server parameters for user groups:
```
    group = groupadmin {
        default service = permit
        enable = permit
        service = shell {
            default command = permit
            default attribute = permit
            set priv-lvl = 15
        }
    }

    group = groupguest {
        default service = permit
        service = shell {
            default command = permit
            default attribute = permit
            set priv-lvl = 1
        }
    }

```

#### config/tac_plus.cfg -> Change the TACACS+ Server parameters for users and passwords:
```
    user = capacita {
        password = clear capacitapass
        member = groupadmin
    }

    user = guest {
        password = clear guestpass
        member = groupadmin
    }
```

### Start Application's Container: 
Run the command below to start docker-compose file: 

```
docker compose up & 
```
The "&" character creates a process id for the command inputed in, with means that the container will not stop when you close the terminal. 

---

### Using Application Server:

Once the container is up and running, your devices can authenticate by TACACS+ connecting to `49/TCP` port, listening to TACACS authentication solicitations. 

--- 
### Stop Container: 
To stop the running container, use the following command:

```
docker-compose down
```

This command stops and removes the containers, networks, defined in the docker-compose.yml file.

--- 

# References/Libs used: 

[Base image (alpine-202104181633) used ](https://hub.docker.com/r/lfkeitel/tacacs_plus)

[TACACS+ documentation for server](http://www.pro-bono-publico.de/projects/unpacked/doc/tac_plus.pdf)
