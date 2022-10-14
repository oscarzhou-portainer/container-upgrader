# container-upgrader
a helper docker image to upgrade the docker container which created it


## Research Process


This fedex project aims to find a way to allow portainer instance (server) to bulk upgrade all agents. 

The expected result will be that a central control page will list each environment(endpoint)'s current version and show if there is new version available. And a bulk upgrade campaign button is on the page to allow users to create a campaign to upgrade agents not in the latest version.

The workflow will be  
1. On UI page, user selects target agents and click upgrade button 
2. API server parse the target agents and broadcast the upgrade request one by one
3. agent receive the upgrade request with necessary information including target agent version or other special commands
4. agent communicate to its containerization environment to schedule the upgrade task. 

During the course, I realized that the agents deployed by ochestrators, such as swarm, kubernetes, and nomad, are able to be redeployed with new version by sending the requests to ochestrators. We just need to figure out what APIs and sdk methods need to be used. However, for docker standalone agent, docker doesn't provide a way for self upgrade. So this project is changed to focus on how to solve docker standalone agent self upgrade. 

WIth the inspiration of unpacker image, a workaround for this issue can be the portainer agent spawns another container, let's say upgrader container, to help finish the upgrade job. The workflow can be  
1. portainer agent needs to start a upgrade container and passing itself docker container ID as an environment variable to it. This can be done by using docker sdk with the `docker.sock` file  
2. After upgrade container run up, it needs to inspect the portainer agent container information, including docker config and current version, and save them for the rollback if posisble
3. upgrade container stops portainer agent container and remove it.
4. upgrade container pulls the portainer agent with the target version and start the container with the same docker config saved from step 2. In addition, passing a flag to instruct portainer agent to delete upgrade container if the upgrading is successful.
5. if step 4 succeeds, new portainer agent needs to remove upgrade container, and communicate to portainer server to update the corresponding environment version in database
6. if step 4 fails, upgrader container needs to rollback to the previous portainer agent version and pass the failure message to it. And portainer agent with the previous version needs to send the failure message back to portainer server.


## Hands-on Expirement


login personal dockerhub

1. create dockerfile 

```
FROM alpine:3.14

WORKDIR /var/fedex
RUN echo "version 1" > ./version.txt
```

2. build and tag ubuntu image to oscarzhou/ubuntu:1

```
docker image build -t oscarzhou/ubuntu:1 .
docker image push oscarzhou/ubuntu:1
```

3. check if the image is as expected

```
docker container run -it oscarzhou/ubuntu:1 /bin/sh

docker container run -it -v /var/run/docker.sock:/var/run/docker.sock oscarzhou/ubuntu:1 /bin/sh
```

4. create dockerfile 2

`touch DockerfileV2`  


```
FROM alpine:3.14

WORKDIR /var/fedex
RUN echo "version 2" > ./version.txt
```

5. build and tag ubuntu image to oscarzhou/ubuntu:2

```
docker image build -t oscarzhou/ubuntu:2 -f DockerfileV2 .
docker image push oscarzhou/ubuntu:2
```

6. check if the image is as expected

```
docker container run -it oscarzhou/ubuntu:2 /bin/sh

docker container run oscarzhou/ubuntu:2 cat version.txt

docker container run -it -v /var/run/docker.sock:/var/run/docker.sock oscarzhou/ubuntu:2 /bin/sh
```

7. Workflow
    1. create container 1
    2. container 1 starts the upgrader container and register the target image and container info(docker run commands) via docker.sock 
    3. upgrader container rm container 1
    4. upgrader container read tasks and start container 2 with flag to tell container 2 to kill upgrader 
    5. container 2 is started and kill upgrader 


## Extra Trick


1. How to get docker container ID inside a container 

`#~ hostname`

2. How to get container configuration

docker container inspect {dockerid}, in container, we can use docker.sock to get it


## Local Environment Debugging


1. tag dev agent docker image and push to private docker hub for using in other docker environment

```
./dev.sh compile
./dev.sh build
docker tag portainerci/agent:develop-8f24881 oscarzhou/agent:fedex-4
docker image push oscarzhou/agent:fedex-4
```

2. Run up fedex-1 in another standalone host

```
docker container rm portainer_agent -f
docker run -d -p 9005:9001 --name portainer_agent --restart=always -v /var/run/docker.sock:/var/run/docker.sock -v /var/lib/docker/volumes:/var/lib/docker/volumes oscarzhou/agent:fedex-4
```