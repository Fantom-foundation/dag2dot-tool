### dot-tool

This util get events data from lachesis node and create DAG on .dot and .png files.

#### Simple start

You should have installed and runing dockerd service.
You should have installed Go lang compiler.

First, start Lachesis nodes:
```bash
cd ~/
git clone https://github.com/Fantom-foundation/go-lachesis.git
cd go-lachesis/docker
# N=<count of nodes required for start>
N=5 ./start.sh

# wait for all nodes run and connected
```

Second, build dot2dot-tool:
```bash
cd ~/
git clone https://github.com/Fantom-foundation/dag2dot-tool.git
cd dag2dot-tool
make
```

Third, run the example:
```bash
# run the script with N=5
./bin/start.sh 5
# outputs are generated in the folder ./lachesis_images

# stop the script
./bin/stop.sh
```

#### Parameters

**-mode** (root|epoch) - mode of create output files. 
In "root" mode separate files created at every changes in root nodes of graph (mean only graph root. not IsRoot status). 
In "epoch" mode separate files created at every epoch. But in process for last epoch files replaced in process creating DAG (possible watch in process).

**-host** - rpc host of lachesis node for requests. Default - "localhost".

**-port** - rpc port of lachesis node for requests. Default - 18545.

**-limit** - for limit count of used events by level, you can use this param. It is usable for very big DAG for watch only top of graph - with changed data.

**-out** - path of directory where will be writing .dot and .png files.

#### Output file names

In "root" mode output file names generated like "DAG{unix nano time}.{dot|png}".

In "epoch" mode output file names generated like "DAG-EPOCH-{epoch number}.{dot|png}"

#### Node information on graph

Node information on graph mean:

* **Triple octagon** - main root (not status - only graph position);
* **Red** border or **red** edges - new elements relative to the previous graph;
* **Yellow** color of node - IsRoot status node (DAG status);

Inside node:
* First line: {epoch}-{lamport time}-{hex 4 bytes of event hash}
* Second line: {frame}-{sequence} (trxs:{count of transactions in event}|)  


#### More details

```bash
cd ~/
git clone https://github.com/Fantom-foundation/dag2dot-tool.git
cd dag2dot-tool
go build ./cmd/dot-tool

# For first node using port 4000, for second - 4001, etc...
mkdir ./images
./dot-tool -mode epoch -host localhost -port 4000 -out ./images
# It is worked in infinity loop mode. When you have enough images, press Ctrl-C on runing dot-tool  
# Watch images in directory ./images

# If you want watch detailing process of creating DAG, use command:
./dot-tool -mode root -host localhost -port 4000 -out ./images
```

#### Build command
``` 
go build cmd/dot-tool/dot-tool.go
```
#### Run example
```./dot-tool -mode epoch -host localhost -port 18546 -out "/tmp/dag"```
