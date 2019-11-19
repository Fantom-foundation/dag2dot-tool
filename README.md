### dot-tool

This util get events data from lachesis node and create DAG on .dot and .png files.

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

#### Build
``` 
go build cmd/dot-tool/dot-tool.go
```
#### Run example
```./dot-tool -mode epoch -host localhost -port 18546 -out "/tmp/dag"```
