# Consecution
Tests of consecutively piping functions over a network
![flow diagram](https://raw.githubusercontent.com/consecution/consecution/master/consecution.png)

## POC
This was a POC to explore the options of using a queue as a replacement for the native kernel pipes.

It wanted to achieve a few goals
- containerized applications that worked only off of StdIn StdOut
- a way to cobble together disparate applications utilizing yaml
- horizontal scalability

## Basics

### Chain file
A chain file is a yaml that declares a series of funcions to be called.  Each function is a link.

Each link is given a unique Id which is just an md5 of the image name, command, and arguments.

```
Chain:
- Image: 'busybox'
  Command: "/bin/cat"
  Name: 'cat'
- Image: 'busybox'
  Command: "/bin/grep"
  Name: 'grep'
  Arguments:
          - 'data'
- Image: 'busybox'
  Command: "/bin/xargs"
  Name: 'xargs'
  Arguments:
          - '-n'
          - '1'

```

### Components

There are 4 required components

#### Portal

The portal serves as an entry point into the system.  In the demo-portal it is using http, but you could use any tool to pass in raw bytes.

```
	etcd := []string{"http://etcd:2379"}
	nats := "nats://nats:4222"
	file := "/files/chain.yaml"
	p, err := portal.New(file, nats, etcd)
	if err != nil {
		log.Fatal(err)
	}
	b := []byte("yourpayload")
	resp, err = p.Send(b)
```
The portal loads at the minimum one chain file, which describes what it expects the backend to do.

The portal works as a controller for the chain file.  For each link in the chain it gets a response until all the links have been processed.

#### Etcd
Etcd just handles holding all the links in the chain.  Multiple portals can publish all their links to the same etcd.  They are published on the link Id, and are not beholden to the portal that created them, or the chain they belong to.  Multiple portals can publish the same link.

#### Nats

Nats is the primary communication channel.  It has a topic for each link Id.  A portal sends a request to the topic that shares the id of its link, and nats queues up a temporary topic for it to receive a reply.

#### Runner

The runner executes each link.  It has no knowledge of where a request comes in.  It works on a QueueSubscribe, where you can have multiple runners subscribed to a topic and only one will get the message.  So as you add more runners you can increase your capacity.

When the runner starts it loads all the links from etcd.

##### Image
Each link in a chain references a Docker Image.  For each link it pulls the docker image, creates a container and then exports that containers filesystem.  Each image only gets one filesystem.

##### Warming
The runner then begins to bootstrap a queue of commands (currently set to 2) for each link.  It will natively containerize the command in the filesystem obtained from the docker image.  These commands are running and waiting for Stdin.  Native containerization is preferred for speed.Running this through the docker daemon (using the demo) was a difference between 1s execution for an entire chain vs 100ms.  

##### Execution
When a request comes in from Nats, the runner finds the command queue by the link Id and pops off one from the queue and takes the input from nats and sends it to Stdin for the command.  In another routine it starts up another command and adds it back to the queue to replace the used one. It collects the reponse from Stdout and sends it back to the reply topic on nats.

## Demo
```
docker-compose build
docker-compose up
```
If everything is up and running

```
time curl -X POST -H "Content-Type: text/plain" --data "this is raw data\n" http://localhost:8080
```

## Footnote
The code base on this project is less than desirable.
